package main

import (
	"context"
	"net/http"
	"supersolik/greed/views"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		err := views.Index("Sergei").Render(context.Background(), c.Response().Writer)

		if err != nil {
			return c.String(http.StatusInternalServerError, "unable to render template")
		}

		return nil
	})
	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
