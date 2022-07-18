package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maize/internal/cards"
	"maize/internal/encryption"
	"maize/internal/models"
	"maize/internal/urlsigner"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminal displays the virtual terminal page
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// TransactionData is the data structure for the transaction
type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

type Invoice struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// GetTransactionData gets the transaction data from the request
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, _ := strconv.Atoi(paymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}
	return txnData, nil
}

// PaymentSucceeded displays the receipt page
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusId: 2,
	}
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// PaymentSucceeded processes a successful payment
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read posted data

	maizeID, err := strconv.Atoi(r.Form.Get("product_id"))
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusId: 2,
	}
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	order := models.Order{
		MaizeID:       maizeID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderID, err := app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	inv := Invoice{
		ID:        orderID,
		Amount:    order.Amount,
		Product:   "Maize",
		Quantity:  order.Quantity,
		FirstName: txnData.FirstName,
		LastName:  txnData.LastName,
		Email:     txnData.Email,
		CreatedAt: time.Now(),
	}

	err = app.callInvoiceMicroService(inv)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

func (app *application) callInvoiceMicroService(inv Invoice) error {
	url := GoDotEnvVariable("INVOICE_SERVICE_URL")
	out, err := json.MarshalIndent(inv, "", "\t")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(out))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	app.infoLog.Println("response Status & response body:", resp.Status, resp.Body)

	return nil
}

// BronzePlanReceipt displays the bronze plan receipt page
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Println(err)
		return
	}
}

// Receipt displays the generic receipt page
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// VirtualTerminalReceipt displays the virtual terminal receipt page
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a customer to the database
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SaveTransaction saves a transaction to the database
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ChargeOnce charges the customer once
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	maizeID, _ := strconv.Atoi(id)

	maize, err := app.DB.GetMaize(maizeID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["maize"] = maize

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// BronzePlan creates a bronze plan for the customer
func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	maize, err := app.DB.GetMaize(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["maize"] = maize
	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, err := app.DB.Authenticate(email, password)
	if err != nil {
		app.errorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	url := r.RequestURI
	testUrl := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}

	valid := signer.VerifyToken(testUrl)

	if !valid {
		app.errorLog.Println("Invalid url tampering detected")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	expired := signer.Expired(testUrl, 60)
	if expired {
		app.errorLog.Println("Url expired")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretkey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["email"] = encryptedEmail

	if err := app.renderTemplate(w, r, "reset-password", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-sales", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) AllSubs(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-subs", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowSale(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Sale"
	stringMap["back"] = "/admin/all-sales"
	stringMap["refund-url"] = "/api/admin/refund"
	stringMap["refund-btn"] = "Refund Order"
	if err := app.renderTemplate(w, r, "sale", &templateData{
		StringMap: stringMap,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowSub(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Subscription"
	stringMap["back"] = "/admin/all-subs"
	stringMap["refund-url"] = "/api/admin/cancel-sub"
	stringMap["refund-btn"] = "Cancel Subscription"
	if err := app.renderTemplate(w, r, "sale", &templateData{
		StringMap: stringMap,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-users", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) OneUser(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "one-user", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}
