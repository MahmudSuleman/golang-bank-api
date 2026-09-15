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

// Register creates a user account.
// @Summary Register
// @Description Register a customer for API access.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} User
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
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

	user, err := h.service.Register(
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

		if errors.Is(err, ErrCustomerNotFound) {
			response.ErrorJSON(
				w,
				http.StatusBadRequest,
				"CUSTOMER_NOT_FOUND",
				"Customer does not exist",
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

// Login authenticates a user.
// @Summary Login
// @Description Authenticate using email and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
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

	loginResponse, err := h.service.Login(r.Context(), request)

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}

		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
		return
	}

	response.JSON(w, http.StatusOK, loginResponse)
}
