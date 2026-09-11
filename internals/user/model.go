package user

type User struct {
	Id           int64  `json:"id"`
	CustomerId   int64  `json:"customer_id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type RegisterRequest struct {
	CustomerId int64  `json:"customer_id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
