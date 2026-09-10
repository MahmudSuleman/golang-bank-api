package account

import (
	"bank-api/internals/customer"
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	customerIDString := chi.URLParam(r, "customerID")

	customerID, err := strconv.ParseInt(
		customerIDString,
		10,
		64,
	)

	if err != nil || customerID <= 0 {
		http.Error(
			w,
			"Invalid customer ID",
			http.StatusBadRequest,
		)
		return
	}
	var request CreateAccountRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
	}

	account, err := h.service.Create(r.Context(), customerID, request)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidAccount) {
			http.Error(w, "Invalid account data", http.StatusInternalServerError)
			return
		}

		if errors.Is(err, customer.ErrCustomerNotFound) {
			http.Error(w, "Invalid account data", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, "invalid customer id", http.StatusBadRequest)
		return
	}

	account, err := h.service.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to retrieve account", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) GetByCustomerId(w http.ResponseWriter, r *http.Request) {
	customerIdString := chi.URLParam(r, "id")

	customerID, err := strconv.ParseInt(customerIdString, 10, 64)

	if err != nil {
		http.Error(w, "invalid customer id", http.StatusBadRequest)
		return
	}

	account, err := h.service.GetByCustomerId(r.Context(), customerID)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}

		http.Error(w, "account not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}
