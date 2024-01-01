package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	greed "supersolik/greed/pkg"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderTempl(c echo.Context, t templ.Component) error {
	return t.Render(context.Background(), c.Response().Writer)
}

func main() {
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

	transactions, err := db.Transactions()

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
	os.Exit(0)

	// e := echo.New()
	// e.Use(middleware.Logger())
	//
	// e.GET("/", func(c echo.Context) error {
	// 	err := renderTempl(c, views.Page(
	// 		views.Accounts(
	// 			greed.ExtractValues(greed.AccountsList),
	// 		),
	// 	))
	//
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, "unable to render template")
	// 	}
	//
	// 	return err
	// })
	//
	// e.GET("/transactions", func(c echo.Context) error {
	// 	err := renderTempl(c, views.Page(
	// 		views.Transactions(
	// 			greed.ExtractValues(greed.TransactionsList),
	// 			greed.ExtractValues(greed.AccountsList),
	// 			greed.CategoriesList,
	// 		),
	// 	))
	//
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, "unable to render template")
	// 	}
	//
	// 	return err
	// })
	//
	// e.GET("/accounts/:id", func(c echo.Context) error {
	// 	accountId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	//
	// 	if err != nil {
	// 		return c.String(http.StatusBadRequest, fmt.Sprintf("Unable to parse %v to account id", accountId))
	// 	}
	//
	// 	var edit bool
	//
	// 	editParam := c.QueryParam("edit")
	//
	// 	if editParam == "true" {
	// 		edit = true
	// 	} else {
	// 		edit = false
	// 	}
	//
	// 	account := greed.AccountsList[accountId]
	//
	// 	err = renderTempl(c, views.Account(account, edit))
	//
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, "unable to render template")
	// 	}
	//
	// 	return err
	// })
	//
	// e.POST("/accounts/:id/save", func(c echo.Context) error {
	// 	accountId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	account := greed.AccountsList[accountId]
	//
	// 	parsedAmount, err := greed.ParseBigFloat(c.FormValue("amount"))
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	account.Update(c.FormValue("account_name"), parsedAmount, c.FormValue("description"))
	// 	greed.AccountsList[account.Id] = account
	//
	// 	err = renderTempl(c, views.Account(account, false))
	//
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, "Unable to render template")
	// 	}
	//
	// 	return err
	// })
	//
	// e.GET("/transactions/:id", func(c echo.Context) error {
	// 	transactionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	//
	// 	if err != nil {
	// 		return c.String(http.StatusBadRequest, fmt.Sprintf("Unable to parse %v to transaction id", transactionId))
	// 	}
	//
	// 	var edit bool
	//
	// 	editParam := c.QueryParam("edit")
	//
	// 	if editParam == "true" {
	// 		edit = true
	// 	} else {
	// 		edit = false
	// 	}
	//
	// 	transaction := greed.TransactionsList[uint64(transactionId)]
	//
	// 	err = renderTempl(c, views.Transaction(transaction, greed.ExtractValues(greed.AccountsList), greed.CategoriesList, edit))
	//
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, "unable to render template")
	// 	}
	//
	// 	return err
	// })
	//
	// e.POST("/transactions/:id/save", func(c echo.Context) error {
	// 	transactionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	transaction := greed.TransactionsList[transactionId]
	//
	// 	formValues, err := c.FormParams()
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	fmt.Println("----transaction form start----")
	// 	for k, v := range formValues {
	// 		fmt.Printf("transaction form %v=%v\n", k, v)
	// 	}
	// 	fmt.Println("----transaction form end----")
	//
	// 	// parse datetime
	// 	date := c.FormValue("date")
	// 	hour := c.FormValue("hours")
	// 	minutes := c.FormValue("minutes")
	// 	location, err := time.LoadLocation(c.FormValue("tz"))
	//
	// 	if err != nil {
	// 		return nil
	// 	}
	//
	// 	// Constructing the time layout
	// 	inputTime := fmt.Sprintf("%s %s:%s", date, hour, minutes)
	//
	// 	// Parsing the input time string
	// 	resultTime, err := time.ParseInLocation(greed.DATETIME_INPUT_LAYOUT, inputTime, location)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	// Print the result time
	// 	fmt.Println("Result Time:", resultTime)
	//
	// 	amount := c.FormValue("amount")
	// 	parsedAmount, _, err := big.ParseFloat(amount, 10, 52, big.ToNearestEven)
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	new_account := c.FormValue("account")
	// 	new_category := c.FormValue("category")
	// 	new_description := c.FormValue("description")
	//
	// 	transaction.Update(new_account, parsedAmount, false, new_category, new_description, resultTime)
	//
	// 	greed.TransactionsList[transaction.Id] = transaction
	//
	// 	return renderTempl(c, views.Transaction(transaction, greed.ExtractValues(greed.AccountsList), greed.CategoriesList, false))
	//
	// })
	// e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
