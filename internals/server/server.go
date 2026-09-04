package server

import (
	"bank-api/internals/customer"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(customerHandler *customer.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler)

	r.Get("/customers", customerHandler.GetAll)
	r.Post("/customers", customerHandler.Create)
	r.Get("/customers/{id}", customerHandler.GetById)

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(`{"status":"ok"}`))
}
