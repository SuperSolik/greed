package greed

import "time"

type Account struct {
	Id          uint
	Name        string
	Amount      float32
	Currency    string
	Description string
}

// repr of Transaction for rendering
type Transaction struct {
	Id          uint
	Account     string
	Amount      float32
	IsExpense   bool
	Category    string
	Timestamp   time.Time
	Description string
}

func (acc *Account) Update(name string, amount float32, description string) {
	acc.Name = name
	acc.Amount = amount
	acc.Description = description
}

func (t *Transaction) Update(account string, amount float32, isExpense bool, category string, description string) {
	t.Account = account
	t.Amount = amount
	t.IsExpense = isExpense
	t.Category = category
	t.Description = description
}

var AccountsList = map[uint]Account{
	1: {Id: 1, Name: "Visa Card", Currency: "RSD", Amount: 10_000},
	2: {Id: 2, Name: "Cash", Currency: "EUR", Amount: 2_000},
}

var TransactionsList = map[uint]Transaction{
	1: {Id: 1, Account: "Visa Card", Amount: 100, IsExpense: true, Category: "Groceries", Timestamp: time.Now(), Description: "some"},
	2: {Id: 2, Account: "Cash", Amount: 50, IsExpense: false, Category: "Salary", Timestamp: time.Now(), Description: "test"},
	3: {Id: 3, Account: "Visa Card", Amount: 210, IsExpense: true, Category: "Groceries", Timestamp: time.Now(), Description: "some"},
	4: {Id: 4, Account: "Cash", Amount: 75, IsExpense: false, Category: "Salary", Timestamp: time.Now(), Description: "test"},
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
