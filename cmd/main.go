package main

import (
	"supersolik/greed/pkg/greed"
	"supersolik/greed/pkg/server"

	"github.com/labstack/gommon/log"
)

func main() {
	db, err := greed.ConnectDb()
	defer db.Close()

	if err != nil {
		log.Fatalf("Failed to connect to db: %v", greed.GetDbUrl())
	}

	e := server.BuildServer(db)

	e.Logger.Fatal(e.Start("127.0.0.1:8080"))
}
