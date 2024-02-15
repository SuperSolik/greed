package server

import (
	"database/sql"
	"net/http"
	"supersolik/greed/pkg/greed"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func createApiEndpoints(e *echo.Echo, db *sql.DB) {
	api := e.Group("/v1")

	api.GET("/categories", func(c echo.Context) error {
		categories, err := greed.GetCategories(db)

		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, categories)
	})
}

func BuildApi(db *sql.DB) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())

	createApiEndpoints(e, db)

	return e
}
