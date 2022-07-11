package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is a wrapper around a sql.DB that provides a few convenience methods
type DBModel struct {
	DB *sql.DB
}

// Models is a collection of DBModel
type Models struct {
	DB DBModel
}

// NewModels creates a new Model
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

// Maize is a model for the maize table
type Maize struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// Order is a model for the orders table
type Order struct {
	ID            int         `json:"id"`
	MaizeID       int         `json:"maize_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Maize         Maize       `json:"maize"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

// Status is a model for the status table
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is a model for the transaction_status table
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is a model for the transactions table
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusId int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is a model for the users table
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is a model for the customers table
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// GetMaize returns a single maize by ID
func (m *DBModel) GetMaize(id int) (Maize, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var maize Maize
	row := m.DB.QueryRowContext(ctx,
		`SELECT
		 id, name, description, inventory_level, price, coalesce(image, ''),is_recurring, plan_id,
	 	 created_at, updated_at
	 	 from 
	 		maize
		 where id = ?`, id)
	err := row.Scan(
		&maize.ID,
		&maize.Name,
		&maize.Description,
		&maize.InventoryLevel,
		&maize.Price,
		&maize.Image,
		&maize.IsRecurring,
		&maize.PlanID,
		&maize.CreatedAt,
		&maize.UpdatedAt)
	if err != nil {
		return maize, err
	}

	return maize, nil
}

// InsertTransaction inserts a new transaction
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO transactions
		 (amount, currency, last_four, bank_return_code, expiry_month, expiry_year,
		 payment_intent, payment_method, transaction_status_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusId,
		time.Now(),
		time.Now())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertOrder inserts a new order
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO orders
		 (maize_id, transaction_id, status_id, quantity, customer_id,
		 amount, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.MaizeID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.CustomerID,
		order.Amount,
		time.Now(),
		time.Now())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertCustomer inserts a new customer
func (m *DBModel) InsertCustomer(c Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO customers
		 (first_name, last_name, email, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		c.FirstName,
		c.LastName,
		c.Email,
		time.Now(),
		time.Now())
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetUserByEmail returns a single user by email
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)
	var u User

	row := m.DB.QueryRowContext(ctx,
		`SELECT
		 	id, first_name, last_name, email, password, created_at, updated_at
		from 
			users
		where email = ?`, email)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var passwordHash string

	row := m.DB.QueryRowContext(ctx, `SELECT id, password FROM users WHERE email = ?`, email)
	err := row.Scan(&id, &passwordHash)
	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("invalid password")
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	UPDATE users
	SET password = ?, updated_at = ?
	WHERE id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, hash, time.Now(), u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetAllOrdersPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	stmt := `
	select 
		o.id, o.maize_id, o.transaction_id, o.customer_id,
		o.status_id, o.quantity, o.amount, o.created_at, o.updated_at,
	    m.id, m.name, t.id, t.amount, t.currency, t.last_four,
	    t.expiry_month, t.expiry_year, t.payment_intent, t.bank_return_code,
	    c.id, c.first_name, c.last_name, c.email
	from 	
		orders o
			left join maize m on (o.maize_id = m.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
	where 
		m.is_recurring = 0	
	order BY
		o.created_at desc
		limit ? offset ?`

	rows, err := m.DB.QueryContext(ctx, stmt, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.MaizeID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Maize.ID,
			&o.Maize.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	stmt = `select count(o.id)
		   	  from orders o
			left join maize m on (o.maize_id = m.id)
			where
				m.is_recurring = 0`

	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, stmt)

	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m *DBModel) GetAllOrders() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*Order

	stmt := `
	select 
		o.id, o.maize_id, o.transaction_id, o.customer_id,
		o.status_id, o.quantity, o.amount, o.created_at, o.updated_at,
	    m.id, m.name, t.id, t.amount, t.currency, t.last_four,
	    t.expiry_month, t.expiry_year, t.payment_intent, t.bank_return_code,
	    c.id, c.first_name, c.last_name, c.email
	from 	
		orders o
			left join maize m on (o.maize_id = m.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
	where 
		m.is_recurring = 0	
	order BY
		o.created_at desc`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.MaizeID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Maize.ID,
			&o.Maize.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	return orders, nil
}

func (m *DBModel) GetAllSubsPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	stmt := `
	select 
		o.id, o.maize_id, o.transaction_id, o.customer_id,
		o.status_id, o.quantity, o.amount, o.created_at, o.updated_at,
	    m.id, m.name, t.id, t.amount, t.currency, t.last_four,
	    t.expiry_month, t.expiry_year, t.payment_intent, t.bank_return_code,
	    c.id, c.first_name, c.last_name, c.email
	from 	
		orders o
			left join maize m on (o.maize_id = m.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
	where 
		m.is_recurring = 1	
	order BY
		o.created_at desc
		limit ? offset ?`

	rows, err := m.DB.QueryContext(ctx, stmt, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.MaizeID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Maize.ID,
			&o.Maize.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	stmt = `select count(o.id)
		   	  from orders o
			left join maize m on (o.maize_id = m.id)
			where
				m.is_recurring = 1`

	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, stmt)

	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m *DBModel) GetAllSubs() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*Order

	stmt := `
	select 
		o.id, o.maize_id, o.transaction_id, o.customer_id,
		o.status_id, o.quantity, o.amount, o.created_at, o.updated_at,
	    m.id, m.name, t.id, t.amount, t.currency, t.last_four,
	    t.expiry_month, t.expiry_year, t.payment_intent, t.bank_return_code,
	    c.id, c.first_name, c.last_name, c.email
	from 	
		orders o
			left join maize m on (o.maize_id = m.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
	where 
		m.is_recurring = 1	
	order BY
		o.created_at desc`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.MaizeID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Maize.ID,
			&o.Maize.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	return orders, nil
}

func (m *DBModel) GetOrderByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o Order

	stmt := `
	select
		o.id, o.maize_id, o.transaction_id, o.customer_id,
		o.status_id, o.quantity, o.amount, o.created_at, o.updated_at,
		m.id, m.name, t.id, t.amount, t.currency, t.last_four,
		t.expiry_month, t.expiry_year, t.payment_intent, t.bank_return_code,
		c.id, c.first_name, c.last_name, c.email
	from
		orders o
			left join maize m on (o.maize_id = m.id)
			left join
				transactions t on (o.transaction_id = t.id)
			left join
				customers c on (o.customer_id = c.id)
		where
			o.id = ?`

	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(
		&o.ID,
		&o.MaizeID,
		&o.TransactionID,
		&o.CustomerID,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Maize.ID,
		&o.Maize.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.LastFour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
	)

	if err != nil {
		return o, err
	}

	return o, nil

}

func (m *DBModel) UpdateOrderStatus(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update orders set status_id = ? where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetAllUsers() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []*User

	stmt := `
	select 
		u.id, u.first_name, u.last_name, u.email, u.created_at, u.updated_at
	from 	
		users u
	order by last_name, first_name`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var u User
		err = rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	return users, nil
}

func (m *DBModel) GetOneUserByID(id int) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User

	stmt := `
	select 
		u.id, u.first_name, u.last_name, u.email, u.created_at, u.updated_at
	from 	
		users u
	where 
		u.id = ?`

	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) EditUser(u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	update users set first_name = ?, last_name = ?, email = ?, updated_at = ? where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email, time.Now(), u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) AddUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	insert into users (first_name, last_name, email, password, created_at, updated_at) values (?, ?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email, hash, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `delete from users where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
