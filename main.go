package main

import (
	"bank-api/internals/account"
	"bank-api/internals/customer"
	"bank-api/internals/database"
	"bank-api/internals/server"
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	databaseURL := "postgres://postgres:postgres@localhost:5432/bank_api?sslmode=disable"

	db, err := database.NewPool(ctx, databaseURL)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	customerRepository := customer.NewPostgresRepository(db)
	customerService := customer.NewService(customerRepository)
	customerHandler := customer.NewHandler(customerService)

	accountRepository := account.NewPostgresRepository(db)
	accountService := account.NewService(accountRepository, customerRepository)
	accountHandler := account.NewHandler(accountService)

	router := server.NewRouter(customerHandler, accountHandler)
	fmt.Println("Bank API running on http://localhost:8080")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("server failed", err)
	}
}
