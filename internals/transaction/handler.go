package transaction

import (
	"bank-api/internals/auth"
	"bank-api/internals/response"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetByAccountId godoc
// @Summary Get account transactions
// @Description Returns the ledger entries for an account owned by the authenticated customer, newest first.
// @Tags Transactions
// @Produce json
// @Param id path int64 true "Account ID"
// @Success 200 {array} Transaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id}/transactions [get]
func (h *Handler) GetByAccountId(w http.ResponseWriter, r *http.Request) {
	customerId, ok := r.Context().Value(auth.CustomerIdKey).(int64)
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	accountId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || accountId <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Invalid account id")
		return
	}

	transactions, err := h.service.GetByAccountId(r.Context(), accountId, customerId)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
			return
		}

		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, transactions)
}
