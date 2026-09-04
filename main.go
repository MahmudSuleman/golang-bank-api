package main

import (
	"fmt"
	"net/http"

	"bank-api/internals/customer"
	"bank-api/internals/server"
)

func main() {
	customerRepository := customer.NewMemoryRepository()
	customerService := customer.NewService(customerRepository)
	customerHandler := customer.NewHandler(customerService)
	router := server.NewRouter(customerHandler)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("server failed", err)
	}
}
