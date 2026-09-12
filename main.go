package main

import (
	"bank-api/internals/account"
	"bank-api/internals/auth"
	"bank-api/internals/customer"
	"bank-api/internals/database"
	"bank-api/internals/server"
	"bank-api/internals/user"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	databaseURL := "postgres://postgres:postgres@localhost:5432/bank_api?sslmode=disable"

	db, err := database.NewPool(ctx, databaseURL)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	jwtManager := auth.NewJWTManager("secret", 15*time.Minute)

	customerRepository := customer.NewPostgresRepository(db)
	customerService := customer.NewService(customerRepository)
	customerHandler := customer.NewHandler(customerService)

	accountRepository := account.NewPostgresRepository(db)
	accountService := account.NewService(accountRepository, customerRepository)
	accountHandler := account.NewHandler(accountService)

	userRepository := user.NewPostgresRepository(db)
	userService := user.NewService(userRepository, customerRepository, jwtManager)
	userHandler := user.NewHandler(userService)

	router := server.NewRouter(customerHandler, accountHandler, userHandler)
	fmt.Println("Bank API running on http://localhost:8080")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("server failed", err)
	}
}
