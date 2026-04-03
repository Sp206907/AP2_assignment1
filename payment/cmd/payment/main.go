package main

import (
	"log"
	"payment/internal/app"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=ernar2026 dbname=paymentdb sslmode=disable"

	if err := app.Run(dsn); err != nil {
		log.Fatal(err)
	}
}
