package customer

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func (r *PostgresRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM customers
			WHERE id = $1
		)
`, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresRepository) GetAll(ctx context.Context) ([]Customer, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, first_name, last_name, email
				FROM customers 
				ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []Customer

	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email)

		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *PostgresRepository) GetById(ctx context.Context, id int64) (Customer, error) {
	var customer Customer

	err := r.db.QueryRow(ctx, `
		SELECT id, first_name, last_name, email
		FROM customers
		WHERE id = $1`,
		id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
	)

	if err != nil {
		return Customer{}, err
	}
	return customer, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	customer Customer,
) (Customer, error) {

	err := r.db.QueryRow(ctx, `
		INSERT INTO customers (
			first_name,
			last_name,
			email
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`,
		customer.FirstName,
		customer.LastName,
		customer.Email,
	).Scan(&customer.ID)

	if err != nil {
		return Customer{}, err
	}

	return customer, nil
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}
