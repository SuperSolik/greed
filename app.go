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

		return renderTempl(c, views.Page(views.Stats()))
	})

	e.GET("/accounts", func(c echo.Context) error {
		accounts, err := greed.GetAccounts(db)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Page(
			views.Accounts(accounts),
		))
	})

	e.GET("/accounts/count", func(c echo.Context) error {
		count, err := greed.CountAccounts(db)

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

		account, err := greed.GetAccountById(db, accountId)

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

		_, err = greed.CreateAccount(db, accountName, parsedAmount, currency, description)

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

		account, err := greed.GetAccountById(db, accountId)

		parsedAmount, err := greed.ParseBigFloat(c.FormValue("amount"))

		if err != nil {
			return err
		}

		account.Name = c.FormValue("account_name")
		account.Amount = parsedAmount
		account.Description = c.FormValue("description")

		_, err = greed.UpdateAccount(db, account)

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

		err = greed.DeleteAccount(db, accountId)

		if err != nil {
			return err
		}

		return renderTempl(c, views.RecountAnchor())
	})

	e.GET("/transactions/content", func(c echo.Context) error {
		var filter greed.TransactionFilter

		pageParam := c.QueryParam("page")
		pageSizeParam := c.QueryParam("size")
		search := c.QueryParam("search")
		dateStart := c.QueryParam("date_start")
		dateEnd := c.QueryParam("date_end")

		// parse page
		if pageParam != "" {
			page, err := strconv.ParseUint(pageParam, 10, 64)

			if err != nil {
				return err
			}

			filter.Page = page
		} else {
			filter.Page = 0
		}

		// parse page size
		if pageSizeParam != "" {
			pageSize, err := strconv.ParseUint(pageSizeParam, 10, 64)

			if err != nil {
				return err
			}

			filter.PageSize = pageSize
		} else {
			filter.PageSize = greed.DefaultPageSize
		}

		if search != "" {
			filter.Search = search
		}

		if dateStart != "" {
			parsedDateStart, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateStart)
			if err != nil {
				return err
			}
			filter.DateStart = parsedDateStart.UTC()
		}

		if dateEnd != "" {
			parsedDateEnd, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateEnd)
			if err != nil {
				return err
			}
			// end date is exclusive in sql, so we need to add 1 day to include the end date itself
			filter.DateEnd = parsedDateEnd.AddDate(0, 0, 1).UTC()
		}

		transactions, err := greed.GetTransactions(db, filter)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Transactions(transactions, filter))
	})

	e.GET("/transactions", func(c echo.Context) error {
		initFilter := greed.TransactionFilterDefault()

		transactions, err := greed.GetTransactions(db, initFilter)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Page(
			views.TransactionsData(transactions, initFilter),
		))
	})

	e.GET("/transactions/count", func(c echo.Context) error {
		count, err := greed.CountTransactions(db)

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

		if _, err := greed.CreateTransactionWithRecalc(
			db,
			greed.Account{Id: accountId, Name: accountData[1]},
			parsedAmount,
			greed.Category{Id: categoryId, Name: categoryData[1]},
			createdAt,
			description,
		); err != nil {
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

		transaction, err := greed.GetTransactionById(db, transactionId)

		if err != nil {
			return err
		}

		if edit {
			accounts, err := greed.GetAccounts(db)
			if err != nil {
				return nil
			}

			categories, err := greed.GetCategories(db)
			if err != nil {
				return nil
			}

			return renderTempl(c, views.TransactionForm(transaction, accounts, categories, false))
		}
		return renderTempl(c, views.Transaction(transaction, templ.Attributes{}))
	})

	e.GET("/transactions/new", func(c echo.Context) error {
		accounts, err := greed.GetAccounts(db)
		if err != nil {
			return nil
		}

		categories, err := greed.GetCategories(db)
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

		err = greed.DeleteTransaction(db, transactionId)

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

		transaction, err := greed.GetTransactionById(db, transactionId)

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

		if _, err := greed.UpdateTransactionWithRecalc(db, transaction); err != nil {
			return err
		}

		return renderTempl(c, views.Transaction(transaction, templ.Attributes{}))
	})

	e.GET("/daterange/input", func(c echo.Context) error {
		rangeType := c.QueryParam("date_range_type")

		now := time.Now().UTC()

		switch rangeType {
		case greed.Today:
			return renderTempl(c, views.DateRangeInput(now, now, true))
		case greed.Last7Days:
			return renderTempl(c, views.DateRangeInput(now.AddDate(0, 0, -6), now, true))
		case greed.ThisWeek:
			diff := [7]int{6, 0, 1, 2, 3, 4, 5}
			return renderTempl(c, views.DateRangeInput(now.AddDate(0, 0, -diff[now.Weekday()]), now, true))
		case greed.Last30Days:
			return renderTempl(c, views.DateRangeInput(now.AddDate(0, 0, -29), now, true))
		case greed.ThisMonth:
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			return renderTempl(c, views.DateRangeInput(startOfMonth, now, true))
		case greed.ThisYear:
			startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
			return renderTempl(c, views.DateRangeInput(startOfYear, now, true))
		case greed.Custom:
			return renderTempl(c, views.DateRangeInput(now, now, false))
		}

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
