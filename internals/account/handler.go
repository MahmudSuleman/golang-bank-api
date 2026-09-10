package account

import (
	"bank-api/internals/customer"
	"bank-api/internals/response"
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

	customerID, err := strconv.ParseInt(customerIDString, 10, 64)

	if err != nil || customerID <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer id")
		return
	}

	var request CreateAccountRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON request body")
		return
	}

	account, err := h.service.Create(r.Context(), customerID, request)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND", "Customer not found")
			return
		}

		if errors.Is(err, ErrInvalidAccount) {
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT", "Invalid account data")
			return
		}

		if errors.Is(err, customer.ErrCustomerNotFound) {
			http.Error(w, "Invalid account data", http.StatusInternalServerError)
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response.JSON(w, http.StatusCreated, account)
}

func (h *Handler) GetByCustomerId(w http.ResponseWriter, r *http.Request) {
	customerIdString := chi.URLParam(r, "id")

	customerID, err := strconv.ParseInt(customerIdString, 10, 64)

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_CUSTOMER_ID",
			"Invalid customer ID",
		)
		return
	}

	accounts, err := h.service.GetByCustomerId(r.Context(), customerID)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND", "Customer not found")
			return
		}

		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response.JSON(w, http.StatusOK, accounts)
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil || id <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID")
		return
	}

	account, err := h.service.GetById(r.Context(), id)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response.JSON(w, http.StatusOK, account)
}
