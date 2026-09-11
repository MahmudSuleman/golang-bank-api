package user

import "context"

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetById(ctx context.Context, id int64) (User, error)
}
