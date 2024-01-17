package tests

import (
	"fmt"
	"math/big"
	"os"
	"supersolik/greed/pkg/greed"
	"time"
)

func test_db() {

	db, err := greed.ConnectDb()

	if err != nil {
		fmt.Printf("Failed to connect to DB: %v", err)
		os.Exit(1)
	}

	accounts, err := db.Accounts()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("---accounts---")
	}

	for i, a := range accounts {
		fmt.Printf("%v: %v\n", i, a)
	}

	fmt.Println("---account non existing---")

	a, err := db.AccountById(200)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", a)
	}

	fmt.Println("---create account---")

	a, err = db.CreateAccount("test account", big.NewFloat(420.69), "USD", "")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", a)
	}

	fmt.Println("---update account---")

	a.Currency = "RSD"
	a.Description = "some new description"
	a.Amount = a.Amount.SetFloat64(69.420)
	a.Name = "BRAND NEW NAME"

	fmt.Printf("%v\n", a)

	rowsUpdated, err := db.UpdateAccount(a)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v, rows updated %v\n", a, rowsUpdated)
	}

	fmt.Println("---transactions---")

	transactions, err := db.Transactions(greed.TransactionFilterDefault())

	if err != nil {
		fmt.Println(err)
	} else {
		for i, t := range transactions {
			fmt.Printf("%v: %v\n", i, t)
		}
	}

	fmt.Println("---categories---")

	categories, err := db.Categories()

	if err != nil {
		fmt.Println(err)
	} else {
		for i, a := range categories {
			fmt.Printf("%v: %v\n", i, a)
		}
	}

	fmt.Println("---account 1---")

	a, err = db.AccountById(1)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", a)
	}

	fmt.Println("---transaction 1---")

	t, err := db.TransactionById(1)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", t)
	}

	fmt.Println("---transaction non existing---")

	t, err = db.TransactionById(200)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", t)
	}

	fmt.Println("---create transaction---")

	t, err = db.CreateTransaction(a, big.NewFloat(123.123), false, &categories[0], time.Now(), "some description")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", t)
	}

	fmt.Println("---update transaction---")

	t.Amount = t.Amount.SetFloat64(40404.500)
	t.Description = "new value"
	t.Category = &categories[1]
	t.IsExpense = true
	t.Account = accounts[2]

	rowsUpdated, err = db.UpdateTransaction(t)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", t)
	}
	os.Exit(0)
}
