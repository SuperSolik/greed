package greed

type Account struct {
	Id          uint
	Name        string
	Amount      float32
	Currency    string
	Description string
}

func (acc *Account) Update(name string, amount float32, description string) {
	acc.Name = name
	acc.Amount = amount
	acc.Description = description
}
