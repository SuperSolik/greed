package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"supersolik/greed/services"
	"supersolik/greed/views"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func renderTempl(c echo.Context, t templ.Component) error {
	return t.Render(context.Background(), c.Response().Writer)
}

var accounts = map[uint]greed.Account{
	1: {Id: 1, Name: "Visa Card", Currency: "RSD", Amount: 10_000},
	2: {Id: 2, Name: "Cash", Currency: "EUR", Amount: 2_000},
}

func ExtractValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, value := range m {
		values = append(values, value)
	}

	return values
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		err := renderTempl(c, views.Index(ExtractValues(accounts)))

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

		account := accounts[uint(accountId)]

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

		account := accounts[uint(accountId)]

		amount := c.FormValue("amount")
		parsedAmount, err := strconv.ParseFloat(amount, 32)

		if err != nil {
			return c.String(http.StatusNotAcceptable, fmt.Sprintf("Provided amount=%v is not valid numeric value", amount))
		}

		account.Update(c.FormValue("account_name"), float32(parsedAmount), c.FormValue("description"))
		accounts[uint(accountId)] = account

		err = renderTempl(c, views.Account(account, false))

		if err != nil {
			return c.String(http.StatusInternalServerError, "Unable to render template")
		}

		return err
	})
	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
