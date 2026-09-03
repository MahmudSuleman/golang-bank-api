package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var nextCustomerID = 3

func main() {
	r := chi.NewRouter()

	r.Get("/health", healthHandler)
	r.Get("/api", apiHandler)

	r.Get("/customers", getCustomers)
	r.Post("/customers", createCustomer)
	r.Get("/customers/{id}", getCustomer)

	fmt.Println("Bank API is running on http://localhost:8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("server failed", err)
	}
}

func createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer Customer

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	customers = append(customers, customer)
	customer.ID = nextCustomerID
	nextCustomerID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)

}

func customerHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		getCustomers(w, r)
	case http.MethodPost:
		createCustomer(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getCustomer(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid customer id", http.StatusBadRequest)
		return
	}

	for _, customer := range customers {
		if customer.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(customer)
			return
		}
	}
	fmt.Fprintln(w, id)
}

func getCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"name": "Bank API", "version": "1.0.0"}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Println("failed to encode response", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Println("response encoded successfully")
	return
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var customers = []Customer{
	{1, "John", "Doe", "john@example.com"},
	{2, "Jane", "Doe", "jane@example.com"},
}
