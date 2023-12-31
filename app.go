package main

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"supersolik/greed/pkg"
	"supersolik/greed/views"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func renderTempl(c echo.Context, t templ.Component) error {
	return t.Render(context.Background(), c.Response().Writer)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		err := renderTempl(c, views.Page(
			views.Accounts(
				greed.ExtractValues(greed.AccountsList),
			),
		))

		if err != nil {
			return c.String(http.StatusInternalServerError, "unable to render template")
		}

		return err
	})

	e.GET("/transactions", func(c echo.Context) error {
		err := renderTempl(c, views.Page(
			views.Transactions(
				greed.ExtractValues(greed.TransactionsList),
				greed.ExtractValues(greed.AccountsList),
				greed.CategoriesList,
			),
		))

		if err != nil {
			return c.String(http.StatusInternalServerError, "unable to render template")
		}

		return err
	})

	e.GET("/accounts/:id", func(c echo.Context) error {
		accountId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Unable to parse %v to account id", accountId))
		}

		var edit bool

		editParam := c.QueryParam("edit")

		if editParam == "true" {
			edit = true
		} else {
			edit = false
		}

		account := greed.AccountsList[uint(accountId)]

		err = renderTempl(c, views.Account(account, edit))

		if err != nil {
			return c.String(http.StatusInternalServerError, "unable to render template")
		}

		return err
	})

	e.POST("/accounts/:id/save", func(c echo.Context) error {
		accountId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Unable to parse %v to account id", accountId))
		}

		account := greed.AccountsList[uint(accountId)]

		parsedAmount, err := greed.ParseBigFloat(c.FormValue("amount"))

		if err != nil {
			return err
		}

		account.Update(c.FormValue("account_name"), parsedAmount, c.FormValue("description"))
		greed.AccountsList[uint(accountId)] = account

		err = renderTempl(c, views.Account(account, false))

		if err != nil {
			return c.String(http.StatusInternalServerError, "Unable to render template")
		}

		return err
	})

	e.GET("/transactions/:id", func(c echo.Context) error {
		transactionId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Unable to parse %v to transaction id", transactionId))
		}

		var edit bool

		editParam := c.QueryParam("edit")

		if editParam == "true" {
			edit = true
		} else {
			edit = false
		}

		transaction := greed.TransactionsList[uint(transactionId)]

		err = renderTempl(c, views.Transaction(transaction, greed.ExtractValues(greed.AccountsList), greed.CategoriesList, edit))

		if err != nil {
			return c.String(http.StatusInternalServerError, "unable to render template")
		}

		return err
	})

	e.POST("/transactions/:id/save", func(c echo.Context) error {
		transactionId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			return err
		}

		transaction := greed.TransactionsList[uint(transactionId)]

		formValues, err := c.FormParams()

		if err != nil {
			return err
		}

		fmt.Println("----transaction form start----")
		for k, v := range formValues {
			fmt.Printf("transaction form %v=%v\n", k, v)
		}
		fmt.Println("----transaction form end----")

		// parse datetime
		date := c.FormValue("date")
		hour := c.FormValue("hours")
		minutes := c.FormValue("minutes")
		location, err := time.LoadLocation(c.FormValue("tz"))

		if err != nil {
			return nil
		}

		// Constructing the time layout
		layout := "2006-01-02 15:04"
		inputTime := fmt.Sprintf("%s %s:%s", date, hour, minutes)

		// Parsing the input time string
		resultTime, err := time.ParseInLocation(layout, inputTime, location)
		if err != nil {
			return err
		}

		// Print the result time
		fmt.Println("Result Time:", resultTime)

		amount := c.FormValue("amount")
		parsedAmount, _, err := big.ParseFloat(amount, 10, 52, big.ToNearestEven)

		if err != nil {
			return err
		}

		new_account := c.FormValue("account")
		new_category := c.FormValue("category")
		new_description := c.FormValue("description")

		transaction.Update(new_account, parsedAmount, false, new_category, new_description, resultTime)

		greed.TransactionsList[transaction.Id] = transaction

		return renderTempl(c, views.Transaction(transaction, greed.ExtractValues(greed.AccountsList), greed.CategoriesList, false))

	})
	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
