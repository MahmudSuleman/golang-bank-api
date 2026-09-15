package auth

import (
	"bank-api/internals/response"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIdKey     contextKey = "user_id"
	CustomerIdKey contextKey = "customer_id"
)

func (j *JWTManager) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.ErrorJSON(w, http.StatusUnauthorized, "INVALID_AUTH_HEADER", "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		claims, err := j.ParseToken(tokenString)

		if err != nil {
			response.ErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey, claims.UserId)
		ctx = context.WithValue(ctx, CustomerIdKey, claims.CustomerId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
