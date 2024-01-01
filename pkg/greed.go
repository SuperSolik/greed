package greed

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type Database struct {
	Handle *sql.DB
}

func GetDbUrl() string {
	url := os.Getenv("DB_URL")

	if url == "" {
		url = "file:///tmp/db.sqlite"
	}
	return url
}

func ConnectDb() (Database, error) {
	url := GetDbUrl()

	var database Database

	db, err := sql.Open("libsql", url)
	if err != nil {
		return database, err
	}

	err = db.Ping()
	if err != nil {
		return database, nil
	}

	fmt.Printf("DB %v connected\n", url)

	database.Handle = db
	return database, nil
}

type Account struct {
	Id          int64
	Name        string
	Amount      *big.Float
	Currency    string
	Description string
}

// repr of Transaction for rendering
type Transaction struct {
	Id          int64
	Account     Account
	Amount      *big.Float
	IsExpense   bool
	Category    *Category // optional category
	CreatedAt   time.Time
	Description string
}

type Category struct {
	Id   int64
	Name string
}

// need this to render the transactions with empty category
func (c *Category) RenderName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (acc *Account) Update(name string, amount *big.Float, description string) {
	acc.Name = name
	acc.Amount = amount
	acc.Description = description
}

func (t *Transaction) Update(account Account, amount *big.Float, isExpense bool, category *Category, description string, createdAt time.Time) {
	t.Account = account
	t.Amount = amount
	t.IsExpense = isExpense
	t.Category = category
	t.Description = description
	t.CreatedAt = createdAt
}

func (d Database) Accounts() ([]Account, error) {
	// An albums slice to hold data from returned rows.
	var accounts []Account

	rows, err := d.Handle.Query("select * from accounts order by id asc")
	if err != nil {
		return nil, fmt.Errorf("fetch accounts failed: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var a Account
		var amount float64
		if err := rows.Scan(&a.Id, &a.Name, &amount, &a.Currency, &a.Description); err != nil {
			return nil, fmt.Errorf("fetch accounts row failed: %v", err)
		}
		// float64 -> bigFloat
		a.Amount = big.NewFloat(amount)
		accounts = append(accounts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during accounts iteration: %v", err)
	}

	return accounts, nil
}

func (d Database) Transactions() ([]Transaction, error) {
	// An albums slice to hold data from returned rows.
	var transactions []Transaction

	query := `
SELECT
    transactions.id AS transaction_id,
    accounts.id AS account_id,
    accounts.name AS account_name,
    transactions.amount,
    transactions.is_expense,
    categories.id AS category_id,
    categories.name AS category_name,
    transactions.created_at,
    transactions.description
FROM
    transactions
JOIN accounts ON transactions.account_id = accounts.id
LEFT JOIN categories ON transactions.category_id = categories.id;
`

	rows, err := d.Handle.Query(query)
	if err != nil {
		return nil, fmt.Errorf("fetch transactions failed: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var t Transaction
		var a Account
		var categoryId sql.NullInt64
		var categoryName sql.NullString

		var amount float64
		var createdAt string
		if err := rows.Scan(&t.Id, &a.Id, &a.Name, &amount, &t.IsExpense, &categoryId, &categoryName, &createdAt, &t.Description); err != nil {
			return nil, fmt.Errorf("fetch transactions row failed: %v", err)
		}
		// float64 -> bigFloat
		t.Amount = big.NewFloat(amount)
		t.Account = a

		parsedCreatedAt, err := time.Parse(DATETIME_DB_LAYOUT, createdAt)

		if err != nil {
			return transactions, err
		}

		t.CreatedAt = parsedCreatedAt

		if categoryId.Valid {
			t.Category = &Category{Id: categoryId.Int64, Name: categoryName.String}
		}

		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during transactions iteration: %v", err)
	}

	return transactions, nil
}

func (d Database) Categories() ([]Category, error) {
	// An albums slice to hold data from returned rows.
	var categories []Category

	rows, err := d.Handle.Query("select * from categories order by id asc")
	if err != nil {
		return nil, fmt.Errorf("fetch categories failed: %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, fmt.Errorf("fetch categories row failed: %v", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during categories iteration: %v", err)
	}

	return categories, nil
}
