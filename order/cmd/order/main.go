package main

import (
	"log"
	"order/internal/app"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=ernar2026 dbname=orderdb sslmode=disable"
	paymentServiceURL := "http://localhost:8081"

	if err := app.Run(dsn, paymentServiceURL); err != nil {
		log.Fatal(err)
	}
}
