package greed

import (
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/labstack/gommon/log"
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

	log.Printf("DB %v connected\n", url)

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
	Category    Category
	CreatedAt   time.Time
	Description string
}

type Category struct {
	Id   int64
	Name string
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

func (d Database) CountAccounts() (int64, error) {
	var count int64

	row := d.Handle.QueryRow("select count(*) from accounts")

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (d Database) CreateAccount(
	name string,
	amount *big.Float,
	currency string,
	description string,
) (Account, error) {
	account := Account{
		Name:        name,
		Amount:      amount,
		Currency:    currency,
		Description: description,
	}

	result, err := d.Handle.Exec(
		"insert into accounts (name, amount, currency, description) values (?, ?, ?, ?)",
		account.Name, account.Amount.String(), account.Currency, account.Description,
	)

	if err != nil {
		return account, fmt.Errorf("failed to create account %v: %v", account, err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return account, fmt.Errorf("failed to get last inserted account id %v: %v", account, err)
	}

	account.Id = id

	return account, nil
}

func (d Database) UpdateAccount(account Account) (int64, error) {
	result, err := d.Handle.Exec(
		"update accounts set name = ?, amount = ?, currency = ?, description = ? where accounts.id = ?",
		account.Name, account.Amount.String(), account.Currency, account.Description, account.Id,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update account %v: %v", account, err)
	}

	rowsUpdated, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("failed to get rows updated when updating account id %v: %v", account, err)
	}

	switch {
	case rowsUpdated == 0:
		return rowsUpdated, fmt.Errorf("update for account %v didn't affect any rows", account)
	case rowsUpdated > 2:
		return rowsUpdated, fmt.Errorf("account %v update affected more than 1 row", account)
	}

	return rowsUpdated, nil
}

func (d Database) AccountById(id int64) (Account, error) {
	// An album to hold data from the returned row.
	a := Account{Id: id}

	var amount float64

	row := d.Handle.QueryRow("select name, amount, currency, description from accounts where id = ?", id)
	if err := row.Scan(&a.Name, &amount, &a.Currency, &a.Description); err != nil {
		return a, err
	}

	a.Amount = big.NewFloat(amount)

	return a, nil
}

func (d Database) DeleteAccount(accountId int64) error {
	result, err := d.Handle.Exec(
		`
		delete from accounts
		where accounts.id = ?
		`,
		accountId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete account %v: %v", accountId, err)
	}
	rowsUpdated, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to get last inserted account id %v: %v", accountId, err)
	}

	switch {
	case rowsUpdated == 0:
		return fmt.Errorf("delete for account %v didn't affect any rows", accountId)
	case rowsUpdated > 2:
		return fmt.Errorf("account %v delete affected more than 1 row", accountId)
	}
	return nil
}

type TransactionFilter struct {
	Page      uint64
	PageSize  uint64
	Search    string
	DateStart time.Time
	DateEnd   time.Time
}

const DefaultPageSize uint64 = 20

func (f TransactionFilter) NextPage() TransactionFilter {
	f.Page += 1
	return f
}

func TransactionFilterDefault() TransactionFilter {
	return TransactionFilter{
		Page:     0,
		PageSize: DefaultPageSize, // TODO: bump
	}
}

func (f TransactionFilter) BuildQueryParams() string {
	var params []string

	params = append(params, fmt.Sprintf("page=%d", f.Page))
	params = append(params, fmt.Sprintf("size=%d", f.PageSize))

	if f.Search != "" {
		params = append(params, fmt.Sprintf("search=%s", f.Search))
	}

	if !f.DateStart.IsZero() {
		params = append(params, fmt.Sprintf("date_start=%s", f.DateStart.UTC().Format(time.DateOnly)))
	}

	if !f.DateEnd.IsZero() {
		params = append(params, fmt.Sprintf("date_end=%s", f.DateEnd.UTC().Format(time.DateOnly)))
	}

	return "?" + strings.Join(params, "&")
}

func (d Database) Transactions(filter TransactionFilter) ([]Transaction, error) {
	query := sq.
		Select(
			"transactions.id as transaction_id",
			"accounts.id as account_id",
			"accounts.name as account_name",
			"transactions.amount as amount",
			"transactions.is_expense as is_expense",
			"categories.id as category_id",
			"categories.name as category_name",
			"transactions.created_at as created_at",
			"transactions.description as desription",
		).
		From("transactions").
		Join("accounts ON transactions.account_id = accounts.id").
		LeftJoin("categories on transactions.category_id = categories.id")

	if filter.Search != "" {
		likeTerm := fmt.Sprint("%", filter.Search, "%")
		query = query.Where(sq.Or{
			sq.Like{"account_name": likeTerm},
			sq.Like{"description": likeTerm},
			sq.Like{"category_name": likeTerm},
		})
	}

	if !filter.DateStart.IsZero() {
		query = query.Where(
			sq.GtOrEq{
				"datetime(transactions.created_at)": filter.DateStart.UTC(),
			},
		)
	}

	if !filter.DateEnd.IsZero() {
		// DateEnd is exclusive
		query = query.Where(
			sq.Lt{
				"datetime(transactions.created_at)": filter.DateEnd.UTC(),
			},
		)
	}

	query = query.OrderBy("datetime(transactions.created_at) desc")

	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize).Offset(filter.Page * filter.PageSize)
	}

	sql, args, err := query.ToSql()

	log.Printf("Transactions query: %v", sql)

	if err != nil {
		return nil, err
	}

	var transactions []Transaction

	rows, err := d.Handle.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch transactions failed: %v", err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var t Transaction
		var a Account
		var c Category

		var amount float64
		var createdAt string
		if err := rows.Scan(&t.Id, &a.Id, &a.Name, &amount, &t.IsExpense, &c.Id, &c.Name, &createdAt, &t.Description); err != nil {
			return nil, fmt.Errorf("fetch transactions row failed: %v", err)
		}
		// float64 -> bigFloat
		t.Amount = big.NewFloat(amount)
		t.Account = a
		t.Category = c

		parsedCreatedAt, err := time.Parse(DATETIME_DB_LAYOUT, createdAt)

		if err != nil {
			return transactions, err
		}

		t.CreatedAt = parsedCreatedAt

		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during transactions iteration: %v", err)
	}

	return transactions, nil
}

func (d Database) CountTransactions() (int64, error) {
	var count int64

	row := d.Handle.QueryRow("select count(*) from transactions")

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
func (d Database) TransactionById(id int64) (Transaction, error) {
	// An album to hold data from the returned row.
	t := Transaction{Id: id}
	var a Account

	var categoryId sql.NullInt64
	var categoryName sql.NullString

	var amount float64
	var createdAt string

	query := `
		select
			accounts.id as account_id,
			accounts.name as account_name,
			transactions.amount,
			transactions.is_expense,
			categories.id as category_id,
			categories.name as category_name,
			transactions.created_at,
			transactions.description
		from
			transactions
		join accounts on transactions.account_id = accounts.id
		left join categories on transactions.category_id = categories.id
		where transactions.id = ?;
	`
	row := d.Handle.QueryRow(query, id)
	if err := row.Scan(&a.Id, &a.Name, &amount, &t.IsExpense, &categoryId, &categoryName, &createdAt, &t.Description); err != nil {
		return t, fmt.Errorf("fetch transactions row failed: %v", err)
	}
	// float64 -> bigFloat
	t.Amount = big.NewFloat(amount)
	t.Account = a

	parsedCreatedAt, err := time.Parse(DATETIME_DB_LAYOUT, createdAt)

	if err != nil {
		return t, err
	}

	t.CreatedAt = parsedCreatedAt

	if categoryId.Valid {
		t.Category = Category{Id: categoryId.Int64, Name: categoryName.String}
	}

	return t, nil
}

func (d Database) CreateTransaction(
	account Account,
	amount *big.Float,
	isExpense bool,
	category Category,
	createdAt time.Time,
	description string,
) (Transaction, error) {
	transaction := Transaction{
		Account:     account,
		Amount:      amount,
		IsExpense:   isExpense,
		Category:    category,
		CreatedAt:   createdAt,
		Description: description,
	}

	result, err := d.Handle.Exec(
		`
		insert into transactions (account_id, amount, is_expense, category_id, created_at, description) 
		values (?, ?, ?, ?, ?, ?)
		`,
		transaction.Account.Id, transaction.Amount.String(), transaction.IsExpense, transaction.Category.Id, transaction.CreatedAt.Format(DATETIME_DB_LAYOUT), transaction.Description,
	)
	if err != nil {
		return transaction, fmt.Errorf("failed to create transaction %v: %v", transaction, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return transaction, fmt.Errorf("failed to get last inserted transaction id %v: %v", transaction, err)
	}

	transaction.Id = id
	return transaction, nil
}

func (d Database) UpdateTransaction(transaction Transaction) (int64, error) {
	result, err := d.Handle.Exec(
		`
		update transactions set account_id = ?, amount = ?, is_expense = ?, category_id = ?, created_at = ?, description = ?
		where transactions.id = ?
		`,
		transaction.Account.Id, transaction.Amount.String(), transaction.IsExpense, transaction.Category.Id, transaction.CreatedAt.Format(DATETIME_DB_LAYOUT), transaction.Description, transaction.Id,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update transaction %v: %v", transaction, err)
	}
	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return rowsUpdated, fmt.Errorf("failed to get last inserted transaction id %v: %v", transaction, err)
	}

	switch {
	case rowsUpdated == 0:
		return rowsUpdated, fmt.Errorf("update for transaction %v didn't affect any rows", transaction)
	case rowsUpdated > 2:
		return rowsUpdated, fmt.Errorf("transaction %v update affected more than 1 row", transaction)
	}
	return rowsUpdated, nil
}

func (d Database) DeleteTransaction(transactionId int64) error {
	result, err := d.Handle.Exec(
		`
		delete from transactions
		where transactions.id = ?
		`,
		transactionId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete transaction %v: %v", transactionId, err)
	}
	rowsUpdated, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to get last inserted transaction id %v: %v", transactionId, err)
	}

	switch {
	case rowsUpdated == 0:
		return fmt.Errorf("delete for transaction %v didn't affect any rows", transactionId)
	case rowsUpdated > 2:
		return fmt.Errorf("transaction %v delete affected more than 1 row", transactionId)
	}
	return nil
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
