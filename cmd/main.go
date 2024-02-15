package main

import (
	"supersolik/greed/pkg/greed"
	"supersolik/greed/pkg/server"

	"github.com/labstack/gommon/log"
)

func main() {
	db, err := greed.ConnectDb()

	if err != nil {
		log.Fatalf("Failed to connect to db: %v", greed.GetDbUrl())
	}

	e := server.BuildWebApp(db)

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
