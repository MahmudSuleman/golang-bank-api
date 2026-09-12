package user

import (
	"bank-api/internals/auth"
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCustomerNotFound   = errors.New("customer not found")
)

type CustomerChecker interface {
	Exists(ctx context.Context, id int64) (bool, error)
}
type Service struct {
	repository      Repository
	customerChecker CustomerChecker
	jwtManage       *auth.JWTManager
}

func NewService(repository Repository, customerChecker CustomerChecker, jwtManage *auth.JWTManager) *Service {
	return &Service{repository: repository, customerChecker: customerChecker, jwtManage: jwtManage}
}

func (s *Service) Register(ctx context.Context, request RegisterRequest) (User, error) {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.CustomerId <= 0 || request.Email == "" || len(request.Password) < 8 {
		return User{}, ErrInvalidUser
	}

	exists, err := s.customerChecker.Exists(ctx, request.CustomerId)
	if err != nil {
		return User{}, err
	}

	if !exists {
		return User{}, ErrCustomerNotFound
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

func (s *Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	user, err := s.repository.GetByEmail(ctx, email)

	if err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManage.GenerateToken(user.Id, user.CustomerId)

	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{User: user, AccessToken: accessToken}, nil

}
