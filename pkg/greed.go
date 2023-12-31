package greed

import (
	"math/big"
	"time"
)

type Account struct {
	Id          uint
	Name        string
	Amount      *big.Float
	Currency    string
	Description string
}

// repr of Transaction for rendering
type Transaction struct {
	Id          uint
	Account     string
	Amount      *big.Float
	IsExpense   bool
	Category    string
	CreatedAt   time.Time
	Description string
}

func (acc *Account) Update(name string, amount *big.Float, description string) {
	acc.Name = name
	acc.Amount = amount
	acc.Description = description
}

func (t *Transaction) Update(account string, amount *big.Float, isExpense bool, category string, description string, createdAt time.Time) {
	t.Account = account
	t.Amount = amount
	t.IsExpense = isExpense
	t.Category = category
	t.Description = description
	t.CreatedAt = createdAt
}

var AccountsList = map[uint]Account{
	1: {Id: 1, Name: "Visa Card", Currency: "RSD", Amount: big.NewFloat(10000)},
	2: {Id: 2, Name: "Cash", Currency: "EUR", Amount: big.NewFloat(2000)},
}

var TransactionsList = map[uint]Transaction{
	1: {Id: 1, Account: "Visa Card", Amount: big.NewFloat(100), IsExpense: true, Category: "Groceries", CreatedAt: time.Now(), Description: "some"},
	2: {Id: 2, Account: "Cash", Amount: big.NewFloat(50), IsExpense: false, Category: "Salary", CreatedAt: time.Now(), Description: "test"},
	3: {Id: 3, Account: "Visa Card", Amount: big.NewFloat(210), IsExpense: true, Category: "Groceries", CreatedAt: time.Now(), Description: "some"},
	4: {Id: 4, Account: "Cash", Amount: big.NewFloat(75), IsExpense: false, Category: "Salary", CreatedAt: time.Now(), Description: "test"},
}

var CategoriesList = []string{
	"Groceries",
	"Salary",
	"Fun",
	"Taxes",
}

func ExtractValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, value := range m {
		values = append(values, value)
	}

	return values
}

func ParseBigFloat(x string) (*big.Float, error) {
	parsedX, _, err := big.ParseFloat(x, 10, 53, big.ToNearestEven)
	return parsedX, err
}
