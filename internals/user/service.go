package user

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, request RegisterRequest) (User, error) {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.CustomerId <= 0 || request.Email == "" || len(request.Password) < 8 {
		return User{}, ErrInvalidUser
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	if err != nil {
		return User{}, err
	}

	var user = User{
		CustomerId:   request.CustomerId,
		Email:        request.Email,
		PasswordHash: string(passwordHash),
	}

	return s.repository.Create(ctx, user)

}

func (s *Service) Login(ctx context.Context, request LoginRequest) (User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	user, err := s.repository.GetByEmail(ctx, email)

	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil


}
