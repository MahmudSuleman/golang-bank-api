package customer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetAll godoc
// @Summary Get all customers
// @Description Retrieves all customers.
// @Tags Customers
// @Produce json
// @Success 200 {array} Customer
// @Failure 400 {string} string
// @Security BearerAuth
// @Router /customers [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve customers", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

// GetById GoDoc
// @Summary Get customer by ID
// @Description Retrieves a customer by their ID.
// @Tags Customers
// @Produce json
// @Param id path int64 true "Customer ID"
// @Success 200 {object} Customer
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Security BearerAuth
// @Router /customers/{id} [get]
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		http.Error(w, "invalid customer id", http.StatusBadRequest)
		return
	}
	customer, err := h.service.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// Create GoDoc
// @Summary Create a customer
// @Description Creates a new customer.
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body CreateCustomerRequest true "Customer details"
// @Success 201 {object} Customer
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Security BearerAuth
// @Router /customers [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

	var request CreateCustomerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	created, err := h.service.Create(r.Context(), request)
	if err != nil {
		if errors.Is(err, ErrInvalidCustomer) {
			http.Error(w, "Invalid customer data", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create customer", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(created)
}
