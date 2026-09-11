package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func (r PostgresRepository) Create(ctx context.Context, user User) (User, error) {

	sqlString := `INSERT INTO users (customer_id,email,password_hash) 
					VALUES ($1, $2, $3) 
					RETURNING id`

	err := r.db.QueryRow(ctx, sqlString, user.CustomerId, user.Email, user.PasswordHash).Scan(&user.Id)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r PostgresRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var user User

	err := r.db.QueryRow(ctx, `SELECT id, customer_id, email, password_hash FROM users WHERE email = $1`, email).
		Scan(&user.Id, &user.CustomerId, &user.Email, &user.PasswordHash)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r PostgresRepository) GetById(ctx context.Context, id int64) (User, error) {
	var user User

	err := r.db.QueryRow(ctx, `SELECT id, customer_id, email, password_hash FROM users WHERE id = $1`, id).
		Scan(&user.Id, &user.CustomerId, &user.Email, &user.PasswordHash)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}
