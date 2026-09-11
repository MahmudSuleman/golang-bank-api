package user

import (
	"bank-api/internals/response"
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"Invalid JSON request body",
		)
		return
	}

	user, err := h.service.Create(
		r.Context(),
		request,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidUser) {
			response.ErrorJSON(
				w,
				http.StatusBadRequest,
				"INVALID_USER",
				"Invalid user data",
			)
			return
		}

		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An internal error occurred",
		)
		return
	}

	response.JSON(
		w,
		http.StatusCreated,
		user,
	)
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request LoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"Invalid JSON request body",
		)
		return
	}

	user, err := h.service.Login(r.Context(), request)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}

		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
