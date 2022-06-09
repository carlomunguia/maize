package card

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusId int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}
	return pi, "", nil
}

func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined."
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired."
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Your card's security code is incorrect."
	case stripe.ErrorCodeIncorrectNumber:
		msg = "Your card number is incorrect."
	case stripe.ErrorCodeIncorrectZip:
		msg = "Your card's zip code failed validation."
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to process."
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to process."
	case stripe.ErrorCodePostalCodeInvalid:
		msg = "Your card's postal code failed validation."
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Your card's balance is insufficient."
	default:
		msg = "There was an error processing your card."
	}

	return msg
}
