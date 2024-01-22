package greed

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type DatabaseInterface interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func GetDbUrl() string {
	url := os.Getenv("DB_URL")

	if url == "" {
		url = "file:///tmp/db.sqlite"
	}
	return url
}

func ConnectDb() (*sql.DB, error) {
	url := GetDbUrl()

	db, err := sql.Open("libsql", url)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Printf("DB %v connected\n", url)

	return db, nil
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
	Category    Category
	CreatedAt   time.Time
	Description string
}

type Category struct {
	Id   int64
	Name string
}

func GetAccounts[T DatabaseInterface](db T) ([]Account, error) {
	// An albums slice to hold data from returned rows.
	var accounts []Account

	rows, err := db.Query("select * from accounts order by id desc")
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

func CountAccounts[T DatabaseInterface](db T) (int64, error) {
	var count int64

	row := db.QueryRow("select count(*) from accounts")

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func CreateAccount[T DatabaseInterface](
	db T,
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

	result, err := db.Exec(
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

func UpdateAccount[T DatabaseInterface](db T, account Account) (int64, error) {
	result, err := db.Exec(
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

func GetAccountById[T DatabaseInterface](db T, id int64) (Account, error) {
	// An album to hold data from the returned row.
	a := Account{Id: id}

	var amount float64

	row := db.QueryRow("select name, amount, currency, description from accounts where id = ?", id)
	if err := row.Scan(&a.Name, &amount, &a.Currency, &a.Description); err != nil {
		return a, err
	}

	a.Amount = big.NewFloat(amount)

	return a, nil
}

func DeleteAccount[T DatabaseInterface](db T, accountId int64) error {
	result, err := db.Exec(
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

type DateRange struct {
	DateStart time.Time
	DateEnd   time.Time
}

type TransactionFilter struct {
	Page          uint64
	PageSize      uint64
	Search        string
	DateRange     DateRange
	FilterExpense bool
	FilterIncome  bool
}

const DefaultPageSize uint64 = 15

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

	if !f.DateRange.DateStart.IsZero() {
		params = append(params, fmt.Sprintf("date_start=%s", f.DateRange.DateStart.UTC().Format(time.DateOnly)))
	}

	if !f.DateRange.DateEnd.IsZero() {
		params = append(params, fmt.Sprintf("date_end=%s", f.DateRange.DateEnd.UTC().Format(time.DateOnly)))
	}

	return "?" + strings.Join(params, "&")
}

func GetTransactions[T DatabaseInterface](db T, filter TransactionFilter) ([]Transaction, error) {
	log.Printf("Querying transactions with filter=%v", filter)

	query := sq.
		Select(
			"transactions.id as transaction_id",
			"accounts.id as account_id",
			"accounts.name as account_name",
			"transactions.amount as amount",
			"categories.id as category_id",
			"categories.name as category_name",
			"transactions.created_at as created_at",
			"transactions.description as transaction_description",
		).
		From("transactions").
		Join("accounts ON transactions.account_id = accounts.id").
		LeftJoin("categories on transactions.category_id = categories.id")

	if filter.FilterExpense {
		query = query.Where(
			sq.Lt{
				"transactions.amount": 0,
			},
		)
	}

	if filter.FilterIncome {
		query = query.Where(
			sq.GtOrEq{
				"transactions.amount": 0,
			},
		)
	}

	if filter.Search != "" {
		likeTerm := fmt.Sprint("%", filter.Search, "%")
		query = query.Where(sq.Or{
			sq.Like{"account_name": likeTerm},
			sq.Like{"transaction_description": likeTerm},
			sq.Like{"category_name": likeTerm},
		})
	}

	if !filter.DateRange.DateStart.IsZero() {
		query = query.Where(
			sq.GtOrEq{
				"datetime(transactions.created_at)": filter.DateRange.DateStart.UTC(),
			},
		)
	}

	if !filter.DateRange.DateEnd.IsZero() {
		// DateRange.DateEnd is exclusive
		query = query.Where(
			sq.Lt{
				"datetime(transactions.created_at)": filter.DateRange.DateEnd.UTC(),
			},
		)
	}

	query = query.OrderBy("datetime(transactions.created_at) desc")

	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize).Offset(filter.Page * filter.PageSize)
	}

	sql, args, err := query.ToSql()

	// log.Printf("Transactions query: %v", sql)

	if err != nil {
		return nil, err
	}

	var transactions []Transaction

	rows, err := db.Query(sql, args...)
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
		if err := rows.Scan(&t.Id, &a.Id, &a.Name, &amount, &c.Id, &c.Name, &createdAt, &t.Description); err != nil {
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

func CountTransactions[T DatabaseInterface](db T) (int64, error) {
	var count int64

	row := db.QueryRow("select count(*) from transactions")

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
func GetTransactionById[T DatabaseInterface](db T, id int64) (Transaction, error) {
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
	row := db.QueryRow(query, id)
	if err := row.Scan(&a.Id, &a.Name, &amount, &categoryId, &categoryName, &createdAt, &t.Description); err != nil {
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

func CreateTransaction[T DatabaseInterface](
	db T,
	account Account,
	amount *big.Float,
	category Category,
	createdAt time.Time,
	description string,
) (Transaction, error) {
	transaction := Transaction{
		Account:     account,
		Amount:      amount,
		Category:    category,
		CreatedAt:   createdAt,
		Description: description,
	}

	result, err := db.Exec(
		`
		insert into transactions (account_id, amount, category_id, created_at, description) 
		values (?, ?, ?, ?, ?)
		`,
		transaction.Account.Id, transaction.Amount.String(), transaction.Category.Id, transaction.CreatedAt.Format(DATETIME_DB_LAYOUT), transaction.Description,
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

func CreateTransactionWithRecalc(
	db *sql.DB,
	account Account,
	amount *big.Float,
	category Category,
	createdAt time.Time,
	description string,
) (Transaction, error) {
	var transaction Transaction

	tx, err := db.Begin()
	if err != nil {
		return transaction, err
	}
	defer tx.Rollback()

	transaction, err = CreateTransaction(
		tx,
		account,
		amount,
		category,
		createdAt,
		description,
	)

	if err != nil {
		return transaction, err
	}

	account, err = GetAccountById(tx, transaction.Account.Id)

	account.Amount.Add(account.Amount, transaction.Amount)

	if _, err := UpdateAccount(tx, account); err != nil {
		return transaction, err
	}

	if err := tx.Commit(); err != nil {
		return transaction, err
	}

	return transaction, nil
}

func UpdateTransaction[T DatabaseInterface](db T, transaction Transaction) (int64, error) {
	result, err := db.Exec(
		`
		update transactions set account_id = ?, amount = ?, category_id = ?, created_at = ?, description = ?
		where transactions.id = ?
		`,
		transaction.Account.Id, transaction.Amount.String(), transaction.Category.Id, transaction.CreatedAt.Format(DATETIME_DB_LAYOUT), transaction.Description, transaction.Id,
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

func UpdateTransactionWithRecalc(db *sql.DB, transaction Transaction) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	oldTransaction, err := GetTransactionById(tx, transaction.Id)

	if err != nil {
		return 0, err

	}

	rowsUpdated, err := UpdateTransaction(
		tx,
		transaction,
	)

	if err != nil {
		return rowsUpdated, err
	}

	account, err := GetAccountById(tx, transaction.Account.Id)

	account.Amount.Sub(account.Amount, oldTransaction.Amount)
	account.Amount.Add(account.Amount, transaction.Amount)

	if _, err := UpdateAccount(tx, account); err != nil {
		return rowsUpdated, err
	}

	if err := tx.Commit(); err != nil {
		return rowsUpdated, err
	}

	return rowsUpdated, nil
}

func DeleteTransaction[T DatabaseInterface](db T, transactionId int64) error {
	result, err := db.Exec(
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

func DeleteTransactionWithRecalc(db *sql.DB, transactionId int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	transaction, err := GetTransactionById(tx, transactionId)

	if err != nil {
		return err
	}

	if err := DeleteTransaction(tx, transactionId); err != nil {
		return err
	}

	account, err := GetAccountById(tx, transaction.Account.Id)
	account.Amount.Sub(account.Amount, transaction.Amount)

	if _, err := UpdateAccount(tx, account); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
func GetCategories[T DatabaseInterface](db T) ([]Category, error) {
	// An albums slice to hold data from returned rows.
	var categories []Category

	rows, err := db.Query("select * from categories order by id asc")
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

type CurrencyAmount struct {
	Currency string
	Amount   *big.Float
}

type CashFlow struct {
	Value    CurrencyAmount
	Positive bool
}

type CategorySpent struct {
	Category Category
	Value    CurrencyAmount
}

type Stats struct {
	Balance         []CurrencyAmount
	CashFlow        []CashFlow
	CategoriesSpent []Pair[string, []CategorySpent]
}

func GetBalance[T DatabaseInterface](db T) ([]CurrencyAmount, error) {

	var result []CurrencyAmount

	sql := `
	select sum(abs(accounts.amount)) as total_amount, accounts.currency as currency
	from accounts 
	group by currency
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("fetch balances failed: %v", err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var ca CurrencyAmount
		var amount float64

		if err := rows.Scan(&amount, &ca.Currency); err != nil {
			return nil, fmt.Errorf("fetch balances row failed: %v", err)
		}

		// float64 -> bigFloat
		ca.Amount = big.NewFloat(amount)

		result = append(result, ca)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during balances iteration: %v", err)
	}

	return result, nil
}

func GetExpensesByCategory[T DatabaseInterface](db T, dateRange DateRange) ([]Pair[string, []CategorySpent], error) {
	var result []Pair[string, []CategorySpent]

	query := sq.
		Select(
			"categories.id",
			"categories.name",
			"sum(abs(transactions.amount)) as total_amount",
			"accounts.currency as currency",
		).
		From("transactions").
		Join("categories on categories.id = transactions.category_id").
		Join("accounts on accounts.id = transactions.account_id").
		Where(sq.Lt{"transactions.amount": 0})

	if !dateRange.DateStart.IsZero() {
		query = query.Where(
			sq.GtOrEq{
				"datetime(transactions.created_at)": dateRange.DateStart.UTC(),
			},
		)
	}

	if !dateRange.DateEnd.IsZero() {
		// DateRange.DateEnd is exclusive
		query = query.Where(
			sq.Lt{
				"datetime(transactions.created_at)": dateRange.DateEnd.UTC(),
			},
		)
	}
	query = query.
		GroupBy("category_id", "currency").
		OrderBy("currency asc")

	sql, args, err := query.ToSql()

	log.Printf("Querying the categories spent with sql: %v", sql)

	if err != nil {
		return nil, err
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch categories total spent failed: %v", err)
	}
	defer rows.Close()

	var prevKey string = ""
	var groupChunk []CategorySpent = nil

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var cs CategorySpent
		var amount float64

		if err := rows.Scan(&cs.Category.Id, &cs.Category.Name, &amount, &cs.Value.Currency); err != nil {
			return nil, fmt.Errorf("fetch categories spent row failed: %v", err)
		}

		// float64 -> bigFloat
		cs.Value.Amount = big.NewFloat(amount)

		key := cs.Value.Currency

		if key != prevKey {
			if len(groupChunk) > 0 {
				sort.SliceStable(groupChunk, func(i, j int) bool {
					return groupChunk[i].Value.Amount.Cmp(groupChunk[j].Value.Amount) > 0
				})

				result = append(result, Pair[string, []CategorySpent]{
					First:  prevKey,
					Second: groupChunk,
				})
				groupChunk = nil
			}
		}

		prevKey = key
		groupChunk = append(groupChunk, cs)
	}

	if len(groupChunk) > 0 {
		sort.SliceStable(groupChunk, func(i, j int) bool {
			return groupChunk[i].Value.Amount.Cmp(groupChunk[j].Value.Amount) > 0
		})
		result = append(result, Pair[string, []CategorySpent]{
			First:  prevKey,
			Second: groupChunk,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during categories spent iteration: %v", err)
	}

	return result, nil
}

func GetCashFlow[T DatabaseInterface](db T, dateRange DateRange) ([]CashFlow, error) {
	query := sq.
		Select(
			"sum(transactions.amount) as cash_flow",
			"accounts.currency as currency",
		).
		From("transactions").
		Join("accounts on accounts.id = transactions.account_id")

	if !dateRange.DateStart.IsZero() {
		query = query.Where(
			sq.GtOrEq{
				"datetime(transactions.created_at)": dateRange.DateStart.UTC(),
			},
		)
	}

	if !dateRange.DateEnd.IsZero() {
		// DateRange.DateEnd is exclusive
		query = query.Where(
			sq.Lt{
				"datetime(transactions.created_at)": dateRange.DateEnd.UTC(),
			},
		)
	}

	query = query.GroupBy("currency")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the cash flow with sql: %v", sql)

	rows, err := db.Query(sql, args...)

	if err != nil {
		return nil, fmt.Errorf("fetch cash flow failed: %v", err)
	}

	var result []CashFlow

	defer rows.Close()
	for rows.Next() {
		var ca CurrencyAmount
		var cashFlow float64

		if err := rows.Scan(&cashFlow, &ca.Currency); err != nil {
			return nil, fmt.Errorf("fetch cash flow row failed: %v", err)
		}

		// float64 -> bigFloat

		ca.Amount = big.NewFloat(cashFlow)

		// check the sign
		isPositive := ca.Amount.Sign() >= 0

		// convert to abs value for repr
		ca.Amount.Abs(ca.Amount)

		result = append(result, CashFlow{Value: ca, Positive: isPositive})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during cash flow iteration: %v", err)
	}

	return result, nil
}

func GetDateRange(rangeType DateRangeType) (DateRange, error) {
	now := time.Now().UTC()

	switch rangeType {
	case Today, Custom:
		return DateRange{now, now}, nil
	case Last7Days:
		return DateRange{now.AddDate(0, 0, -6), now}, nil
	case ThisWeek:
		diff := [7]int{6, 0, 1, 2, 3, 4, 5}
		return DateRange{now.AddDate(0, 0, -diff[now.Weekday()]), now}, nil
	case Last30Days:
		return DateRange{now.AddDate(0, 0, -29), now}, nil
	case ThisMonth:
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return DateRange{startOfMonth, now}, nil
	case ThisYear:
		startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return DateRange{startOfYear, now}, nil
	}

	return DateRange{}, errors.New("unexpected date range type")
}
