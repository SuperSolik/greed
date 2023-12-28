package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"supersolik/greed/pkg"
	"supersolik/greed/views"

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

		amount := c.FormValue("amount")
		parsedAmount, err := strconv.ParseFloat(amount, 32)

		if err != nil {
			return c.String(http.StatusNotAcceptable, fmt.Sprintf("Provided amount=%v is not valid numeric value", amount))
		}

		account.Update(c.FormValue("account_name"), float32(parsedAmount), c.FormValue("description"))
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
		return c.String(http.StatusBadRequest, fmt.Sprintf("TODO ME LATER"))
	})

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
