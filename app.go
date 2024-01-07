package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	greed "supersolik/greed/pkg"
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
	db, err := greed.ConnectDb()

	if err != nil {
		log.Fatalf("Failed to connect to db: %v", greed.GetDbUrl())
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		accounts, err := db.Accounts()

		if err != nil {
			return err
		}

		return renderTempl(c, views.Page(
			views.Accounts(accounts),
		))
	})

	e.GET("/accounts/count", func(c echo.Context) error {
		count, err := db.CountAccounts()

		if err != nil {
			return err
		}

		return c.String(http.StatusOK, strconv.FormatInt(count, 10))
	})

	e.GET("/accounts/:id", func(c echo.Context) error {
		accountId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		var edit bool

		editParam := c.QueryParam("edit")

		if editParam == "true" {
			edit = true
		} else {
			edit = false
		}

		account, err := db.AccountById(accountId)

		if err != nil {
			return err
		}

		if edit {
			return renderTempl(c, views.AccountEdit(account))
		}

		return renderTempl(c, views.Account(account))
	})

	e.POST("/accounts", func(c echo.Context) error {
		accountName := c.FormValue("account_name")
		currency := c.FormValue("currency")
		description := c.FormValue("description")
		parsedAmount, err := greed.ParseBigFloat(c.FormValue("amount"))

		if err != nil {
			return err
		}

		_, err = db.CreateAccount(accountName, parsedAmount, currency, description)

		if err != nil {
			return err
		}

		return renderTempl(c, views.RefreshAnchor())
	})

	e.PUT("/accounts/:id", func(c echo.Context) error {
		accountId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		account, err := db.AccountById(accountId)

		parsedAmount, err := greed.ParseBigFloat(c.FormValue("amount"))

		if err != nil {
			return err
		}

		account.Name = c.FormValue("account_name")
		account.Amount = parsedAmount
		account.Description = c.FormValue("description")

		_, err = db.UpdateAccount(account)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Account(account))
	})

	e.DELETE("/accounts/:id", func(c echo.Context) error {
		accountId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		err = db.DeleteAccount(accountId)

		if err != nil {
			return err
		}

		return renderTempl(c, views.RecountAnchor())
	})

	e.GET("/transactions/content", func(c echo.Context) error {
		var filter greed.TransactionFilter

		// parse page
		var page uint64

		pageParam := c.QueryParam("page")

		if pageParam != "" {
			page, err = strconv.ParseUint(pageParam, 10, 64)

			if err != nil {
				return err
			}
		} else {
			page = 0
		}

		var pageSize uint64

		// parse page size
		pageSizeParam := c.QueryParam("size")

		if pageSizeParam != "" {
			pageSize, err = strconv.ParseUint(pageSizeParam, 10, 64)

			if err != nil {
				return err
			}
		} else {
			pageSize = greed.DefaultPageSize
		}

		filter.Page = page
		filter.PageSize = pageSize

		transactions, err := db.Transactions(filter)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Transactions(transactions, filter))
	})

	e.GET("/transactions", func(c echo.Context) error {
		initFilter := greed.TransactionFilterDefault()

		transactions, err := db.Transactions(initFilter)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Page(
			views.TransactionsData(transactions, initFilter),
		))
	})

	e.GET("/transactions/count", func(c echo.Context) error {
		count, err := db.CountTransactions()

		if err != nil {
			return err
		}

		return c.String(http.StatusOK, strconv.FormatInt(count, 10))
	})

	e.POST("/transactions", func(c echo.Context) error {
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
		inputDate := c.FormValue("date")
		inputTime := c.FormValue("time")

		location, err := time.LoadLocation(c.FormValue("tz"))

		if err != nil {
			return nil
		}

		// Constructing the time layout
		inputDateTime := fmt.Sprintf("%s %s", inputDate, inputTime)

		// Parsing the input time string
		createdAt, err := time.ParseInLocation(greed.DATETIME_INPUT_LAYOUT, inputDateTime, location)

		if err != nil {
			return err
		}

		log.Printf("Parsed  time: %v\n", createdAt)

		amount := c.FormValue("amount")
		parsedAmount, _, err := big.ParseFloat(amount, 10, 53, big.ToNearestEven)

		if err != nil {
			return err
		}

		accountData := strings.Split(c.FormValue("account"), ";")
		if len(accountData) != 2 {
			log.Printf("[ERROR] failed to parse account data: %s", c.FormValue("account"))
			return fmt.Errorf("Error during parsing  account data")
		}

		categoryData := strings.Split(c.FormValue("category"), ";")
		if len(categoryData) != 2 {
			log.Printf("Failed to parse category data: %s", c.FormValue("category"))
			return fmt.Errorf("Error during parsing  category data")
		}

		accountId, err := strconv.ParseInt(accountData[0], 10, 64)
		if err != nil {
			return err
		}

		categoryId, err := strconv.ParseInt(categoryData[0], 10, 64)
		if err != nil {
			return err
		}

		description := c.FormValue("description")

		_, err = db.CreateTransaction(
			greed.Account{Id: accountId, Name: accountData[1]},
			parsedAmount,
			false,
			greed.Category{Id: categoryId, Name: categoryData[1]},
			createdAt,
			description,
		)

		if err != nil {
			return err
		}

		return renderTempl(c, views.RefreshAnchor())
	})

	e.GET("/transactions/:id", func(c echo.Context) error {
		transactionId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		var edit bool

		editParam := c.QueryParam("edit")

		if editParam == "true" {
			edit = true
		} else {
			edit = false
		}

		transaction, err := db.TransactionById(transactionId)

		if err != nil {
			return err
		}

		if edit {
			accounts, err := db.Accounts()
			if err != nil {
				return nil
			}

			categories, err := db.Categories()
			if err != nil {
				return nil
			}

			return renderTempl(c, views.TransactionForm(transaction, accounts, categories, false))
		}
		return renderTempl(c, views.Transaction(transaction, templ.Attributes{}))
	})

	e.GET("/transactions/new", func(c echo.Context) error {
		accounts, err := db.Accounts()
		if err != nil {
			return nil
		}

		categories, err := db.Categories()
		if err != nil {
			return nil
		}

		t := greed.Transaction{
			Category:  categories[0],
			Account:   accounts[0],
			Amount:    big.NewFloat(0.0),
			CreatedAt: time.Now().UTC(),
		}

		return renderTempl(c, views.TransactionForm(t, accounts, categories, true))
	})

	e.DELETE("/transactions/:id", func(c echo.Context) error {
		transactionId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		err = db.DeleteTransaction(transactionId)

		if err != nil {
			return err
		}

		return renderTempl(c, views.RecountAnchor())
	})

	e.PUT("/transactions/:id", func(c echo.Context) error {
		transactionId, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return err
		}

		transaction, err := db.TransactionById(transactionId)

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
		inputDate := c.FormValue("date")
		inputTime := c.FormValue("time")

		location, err := time.LoadLocation(c.FormValue("tz"))

		if err != nil {
			return nil
		}

		// Constructing the time layout
		inputDateTime := fmt.Sprintf("%s %s", inputDate, inputTime)

		// Parsing the input time string
		newCreatedAt, err := time.ParseInLocation(greed.DATETIME_INPUT_LAYOUT, inputDateTime, location)

		if err != nil {
			return err
		}

		log.Printf("Parsed new time: %v\n", newCreatedAt)

		amount := c.FormValue("amount")
		parsedAmount, _, err := big.ParseFloat(amount, 10, 53, big.ToNearestEven)

		if err != nil {
			return err
		}

		newAccountData := strings.Split(c.FormValue("account"), ";")
		if len(newAccountData) != 2 {
			log.Printf("[ERROR] failed to parse account data: %s", c.FormValue("account"))
			return fmt.Errorf("Error during parsing new account data")
		}

		newCategoryData := strings.Split(c.FormValue("category"), ";")
		if len(newCategoryData) != 2 {
			log.Printf("Failed to parse category data: %s", c.FormValue("category"))
			return fmt.Errorf("Error during parsing new category data")
		}

		newAccountId, err := strconv.ParseInt(newAccountData[0], 10, 64)
		if err != nil {
			return err
		}

		newCategoryId, err := strconv.ParseInt(newCategoryData[0], 10, 64)
		if err != nil {
			return err
		}

		newDescription := c.FormValue("description")

		transaction.Amount = parsedAmount
		transaction.Description = newDescription
		transaction.Account = greed.Account{Id: newAccountId, Name: newAccountData[1]}
		transaction.Category = greed.Category{Id: newCategoryId, Name: newCategoryData[1]}
		transaction.CreatedAt = newCreatedAt

		db.UpdateTransaction(transaction)

		return renderTempl(c, views.Transaction(transaction, templ.Attributes{}))
	})
	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
