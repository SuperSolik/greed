package server

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"supersolik/greed/pkg/greed"
	"supersolik/greed/pkg/views"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func renderTempl(c echo.Context, t templ.Component) error {
	return t.Render(context.Background(), c.Response().Writer)
}

func createWebAppEndpoints(e *echo.Echo, db *sql.DB) {
	e.GET("/", func(c echo.Context) error {
		var stats greed.Stats
		defaultRangeType := greed.Last30Days
		defaultDateRange, err := greed.GetDateRange(defaultRangeType)
		if err != nil {
			return err
		}

		if categoriesSpent, err := greed.GetExpensesByCategory(db, defaultDateRange); err != nil {
			return err
		} else {
			stats.CategoriesSpent = categoriesSpent
		}

		if cashFlow, err := greed.GetCashFlow(db, defaultDateRange); err != nil {
			return err
		} else {
			stats.CashFlow = cashFlow
		}

		if balance, err := greed.GetBalance(db); err != nil {
			return err
		} else {
			stats.Balance = balance
		}

		return renderTempl(c, views.Page(views.StatsContent(stats, defaultRangeType)))
	})

	e.GET("/stats/categories", func(c echo.Context) error {
		dateStart := c.QueryParam("date_start")
		dateEnd := c.QueryParam("date_end")

		var dateRange greed.DateRange

		if dateStart != "" {
			parsedDateStart, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateStart)
			if err != nil {
				return err
			}
			dateRange.DateStart = parsedDateStart.UTC()
		}

		if dateEnd != "" {
			parsedDateEnd, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateEnd)
			if err != nil {
				return err
			}
			// end date is exclusive in sql, so we need to add 1 day to include the end date itself
			dateRange.DateEnd = parsedDateEnd.AddDate(0, 0, 1).UTC()
		}

		categoriesSpent, err := greed.GetExpensesByCategory(db, dateRange)

		if err != nil {
			return err
		}

		return renderTempl(c, views.CategoriesExpenses(categoriesSpent))
	})

	e.GET("/stats/cashflow", func(c echo.Context) error {
		dateStart := c.QueryParam("date_start")
		dateEnd := c.QueryParam("date_end")

		var dateRange greed.DateRange

		if dateStart != "" {
			parsedDateStart, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateStart)
			if err != nil {
				return err
			}
			dateRange.DateStart = parsedDateStart.UTC()
		}

		if dateEnd != "" {
			parsedDateEnd, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateEnd)
			if err != nil {
				return err
			}
			// end date is exclusive in sql, so we need to add 1 day to include the end date itself
			dateRange.DateEnd = parsedDateEnd.AddDate(0, 0, 1).UTC()
		}

		if cashFlow, err := greed.GetCashFlow(db, dateRange); err != nil {
			return err
		} else {
			return renderTempl(c, views.CashFlow(cashFlow))
		}
	})

	e.GET("/accounts", func(c echo.Context) error {
		accounts, err := greed.GetAccounts(db)

		if err != nil {
			return err
		}

		return renderTempl(c, views.Page(
			views.AccountsContent(accounts),
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
			return renderTempl(c, views.AccountForm(account, false))
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

		if account, err := greed.CreateAccount(db, accountName, parsedAmount, currency, description); err != nil {
			return err
		} else {
			return renderTempl(c, views.Account(account))
		}
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

	e.GET("/accounts/new", func(c echo.Context) error {
		return renderTempl(c, views.AccountForm(greed.Account{}, true))
	})

	e.GET("/transactions/content", func(c echo.Context) error {
		var filter greed.TransactionFilter

		pageParam := c.QueryParam("page")
		pageSizeParam := c.QueryParam("size")
		search := c.QueryParam("search")
		dateStart := c.QueryParam("date_start")
		dateEnd := c.QueryParam("date_end")
		filterExpense := c.QueryParam("expense") == "true"
		filterIncome := c.QueryParam("income") == "true"

		if filterExpense != filterIncome {
			// either one of them provided  (both not empty, both not true)
			filter.FilterIncome = filterIncome
			filter.FilterExpense = filterExpense
		}

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
			filter.DateRange.DateStart = parsedDateStart.UTC()
		}

		if dateEnd != "" {
			parsedDateEnd, err := time.Parse(greed.DATE_INPUT_LAYOUT, dateEnd)
			if err != nil {
				return err
			}
			// end date is exclusive in sql, so we need to add 1 day to include the end date itself
			filter.DateRange.DateEnd = parsedDateEnd.AddDate(0, 0, 1).UTC()
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
			views.TransactionsContent(transactions, initFilter),
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

		if err = greed.DeleteTransactionWithRecalc(db, transactionId); err != nil {
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
		rangeType := greed.DateRangeType(c.QueryParam("date_range_type"))

		if dateRange, err := greed.GetDateRange(rangeType); err != nil {
			return c.NoContent(http.StatusOK)
		} else {
			return renderTempl(c, views.DateRangeInput(dateRange, rangeType != greed.Custom))
		}
	})
}

func BuildWebApp(db *sql.DB) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())

	createWebAppEndpoints(e, db)

	return e
}
