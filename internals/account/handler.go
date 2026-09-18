package account

import (
	"bank-api/internals/auth"
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

// Withdraw godoc
// @Summary Withdraw money from an account
// @Description Withdraws the specified amount from the account.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path int64 true "Account ID"
// @Param request body MoneyRequest true "Withdrawal amount"
// @Success 200 {object} Account
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id}/withdraw [post]
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	accountIdString := chi.URLParam(r, "id")
	accountId, err := strconv.ParseInt(accountIdString, 10, 64)

	if err != nil || accountId <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Invalid account id")
		return
	}

	var request MoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	account, err := h.service.Withdraw(r.Context(), accountId, request)
	if err != nil {

		switch {
		case errors.Is(err, ErrInvalidAccount):
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT", "Invalid account")

		case errors.Is(err, ErrInvalidAmount):
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_AMOUNT", "Amount must be greater than zero")

		case errors.Is(err, ErrAccountNotFound):
			response.ErrorJSON(w, http.StatusBadRequest, "ACCOUNT_NOT_FOUND", "Account not found")

		case errors.Is(err, ErrInsufficientBalance):
			response.ErrorJSON(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE", "Insufficient balance")
		case errors.Is(err, ErrAccountBlocked):
			response.ErrorJSON(
				w,
				http.StatusForbidden,
				"ACCOUNT_BLOCKED",
				"Account is blocked",
			)

		case errors.Is(err, ErrAccountClosed):
			response.ErrorJSON(w, http.StatusForbidden, "ACCOUNT_CLOSED", "Account is closed")

		default:
			response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}

	}

	response.JSON(w, http.StatusOK, account)
}

// Deposit
// @Summary Deposit money into an account
// @Description Deposits the specified amount into the given account.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path int64 true "Account ID"
// @Param request body MoneyRequest true "Deposit amount"
// @Success 200 {object} Account
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id}/deposit [post]
func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	accountIdString := chi.URLParam(r, "id")
	accountId, err := strconv.ParseInt(accountIdString, 10, 64)

	if err != nil || accountId <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Invalid account id")
	}

	var request MoneyRequest

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	account, err := h.service.Deposit(r.Context(), accountId, request)
	if err != nil {

		switch {
		case errors.Is(err, ErrInvalidAccount):
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT", "Invalid account")
		case errors.Is(err, ErrInvalidAmount):
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_AMOUNT", "Amount must be greater than zero")
		case errors.Is(err, ErrAccountNotFound):
			response.ErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		case errors.Is(err, ErrAccountBlocked):
			response.ErrorJSON(w, http.StatusForbidden, "ACCOUNT_BLOCKED", "Account is blocked")
		case errors.Is(err, ErrAccountClosed):
			response.ErrorJSON(w, http.StatusForbidden, "ACCOUNT_CLOSED", "Account is closed")
		default:
			response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, account)
}

// Create GoDoc
// @Summary Create a bank account
// @Description Creates a new bank account for the authenticated customer.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param customerID path int64 true "Customer ID"
// @Param request body CreateAccountRequest true "Account details"
// @Success 201 {object} Account
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /customers/{customerID}/accounts [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {

	authenticatedCustomerId, ok := r.Context().Value(auth.CustomerIdKey).(int64)
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	customerIDString := chi.URLParam(r, "customerID")

	customerID, err := strconv.ParseInt(customerIDString, 10, 64)

	if err != nil || customerID != authenticatedCustomerId {
		response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You cannot create an account for another customer")
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

// GetByCustomerId GoDoc
// @Summary Get customer accounts
// @Description Returns all bank accounts belonging to the authenticated customer.
// @Tags Accounts
// @Produce json
// @Param id path int64 true "Customer ID"
// @Success 200 {array} Account
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /customers/{id}/accounts [get]
func (h *Handler) GetByCustomerId(w http.ResponseWriter, r *http.Request) {

	authenticatedCustomerId, ok := r.Context().Value(auth.CustomerIdKey).(int64)
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

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

	if authenticatedCustomerId != customerID {
		response.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "You cannot get accounts for another customer")
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

// GetById retrieves an account.
// @Summary Get account
// @Description Get a bank account by ID.
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Account ID"
// @Success 200 {object} Account
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id} [get]
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil || id <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID")
		return
	}

	customerId, ok := r.Context().Value(auth.CustomerIdKey).(int64)
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	account, err := h.service.GetByIdForCustomer(r.Context(), id, customerId)
	if err != nil {
		response.ErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response.JSON(w, http.StatusOK, account)
}
