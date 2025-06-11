package main

import (
	"log"
	"os"

	"github.com/go-telegram/miniapp/server"
)

func main() {
	cfg := server.Config{
		ListenAddress: "3000",
	}

	if v := os.Getenv("LISTEN_ADDRESS"); v != "" {
		cfg.ListenAddress = v
	}
	app := server.NewTelegramApplication()
	router := app.WebsiteRoutes()

	err := app.ServeHTTP(router, 3000)
	if err != nil {
		log.Printf("error run: %v", err)
		os.Exit(1)
	}

}
