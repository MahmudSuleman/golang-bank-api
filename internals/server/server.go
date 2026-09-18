package server

import (
	"bank-api/internals/account"
	"bank-api/internals/auth"
	"bank-api/internals/customer"
	"bank-api/internals/user"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(customerHandler *customer.Handler,
	accountHandler *account.Handler,
	userHandler *user.Handler,
	jwtManager *auth.JWTManager,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))

	r.Post("/auth/register", userHandler.Register)
	r.Post("/auth/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(jwtManager.Authenticate)
		r.Get("/customers", customerHandler.GetAll)
		r.Post("/customers", customerHandler.Create)
		r.Get("/customers/{id}", customerHandler.GetById)
		r.Get("/customers/{id}/accounts", accountHandler.GetByCustomerId)
		r.Post("/customers/{id}/accounts", accountHandler.Create)
		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", accountHandler.Create)
			r.Get("/", accountHandler.GetByCustomerId)
			r.Get("/{id}", accountHandler.GetById)
			r.Post("/{id}/deposit", accountHandler.Deposit)
			r.Post("/{id}/withdraw", accountHandler.Withdraw)
			r.Post("/{id}/transfer", accountHandler.Transfer)
		})

	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(`{"status":"ok"}`))
}
