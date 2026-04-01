package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrServiceNotFound = errors.New("service not found")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]Service, error) {
	query := `
		SELECT
			service_id,
			name,
			COALESCE(description, ''),
			price_rub::float8,
			regular_discount_percent::float8,
			discounted_price_rub::float8
		FROM automaster.services
		ORDER BY service_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]Service, 0)

	for rows.Next() {
		var service Service

		err := rows.Scan(
			&service.ServiceID,
			&service.Name,
			&service.Description,
			&service.PriceRub,
			&service.RegularDiscountPercent,
			&service.DiscountedPriceRub,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repository) Create(ctx context.Context, input CreateServiceInput) (Service, error) {
	query := `
		INSERT INTO automaster.services (name, description, price_rub, regular_discount_percent)
		VALUES ($1, NULLIF($2, ''), $3, $4)
		RETURNING
			service_id,
			name,
			COALESCE(description, ''),
			price_rub::float8,
			regular_discount_percent::float8,
			discounted_price_rub::float8
	`

	var service Service

	err := r.DB.QueryRow(
		ctx,
		query,
		input.Name,
		input.Description,
		input.PriceRub,
		input.RegularDiscountPercent,
	).Scan(
		&service.ServiceID,
		&service.Name,
		&service.Description,
		&service.PriceRub,
		&service.RegularDiscountPercent,
		&service.DiscountedPriceRub,
	)
	if err != nil {
		return Service{}, err
	}

	return service, nil
}

func (r *Repository) Update(ctx context.Context, serviceID int64, input UpdateServiceInput) (Service, error) {
	query := `
		UPDATE automaster.services
		SET
			name = $2,
			description = NULLIF($3, ''),
			price_rub = $4,
			regular_discount_percent = $5
		WHERE service_id = $1
		RETURNING
			service_id,
			name,
			COALESCE(description, ''),
			price_rub::float8,
			regular_discount_percent::float8,
			discounted_price_rub::float8
	`

	var service Service

	err := r.DB.QueryRow(
		ctx,
		query,
		serviceID,
		input.Name,
		input.Description,
		input.PriceRub,
		input.RegularDiscountPercent,
	).Scan(
		&service.ServiceID,
		&service.Name,
		&service.Description,
		&service.PriceRub,
		&service.RegularDiscountPercent,
		&service.DiscountedPriceRub,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, ErrServiceNotFound
		}

		return Service{}, err
	}

	return service, nil
}

func (r *Repository) Delete(ctx context.Context, serviceID int64) error {
	query := `
		DELETE FROM automaster.services
		WHERE service_id = $1
	`

	result, err := r.DB.Exec(ctx, query, serviceID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrServiceNotFound
	}

	return nil
}
