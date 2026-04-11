package maintenanceparts

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMaintenancePartNotFound = errors.New("maintenance part not found")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]MaintenancePart, error) {
	query := `
		SELECT
			maintenance_part_id,
			name,
			quantity,
			COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), '') AS delivery_date
		FROM automaster.maintenance_parts
		ORDER BY maintenance_part_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parts := make([]MaintenancePart, 0)

	for rows.Next() {
		var part MaintenancePart

		err := rows.Scan(
			&part.MaintenancePartID,
			&part.Name,
			&part.Quantity,
			&part.DeliveryDate,
		)
		if err != nil {
			return nil, err
		}

		parts = append(parts, part)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

func (r *Repository) Create(ctx context.Context, input CreateMaintenancePartInput) (MaintenancePart, error) {
	query := `
		INSERT INTO automaster.maintenance_parts (name, quantity, delivery_date)
		VALUES ($1, $2, NULLIF($3, '')::date)
		RETURNING
			maintenance_part_id,
			name,
			quantity,
			COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), '') AS delivery_date
	`

	var part MaintenancePart

	err := r.DB.QueryRow(
		ctx,
		query,
		input.Name,
		input.Quantity,
		input.DeliveryDate,
	).Scan(
		&part.MaintenancePartID,
		&part.Name,
		&part.Quantity,
		&part.DeliveryDate,
	)
	if err != nil {
		return MaintenancePart{}, err
	}

	return part, nil
}

func (r *Repository) Restock(ctx context.Context, partID int64, input RestockMaintenancePartInput) (MaintenancePart, error) {
	query := `
		UPDATE automaster.maintenance_parts
		SET
			quantity = quantity + $2,
			delivery_date = NULLIF($3, '')::date
		WHERE maintenance_part_id = $1
		RETURNING
			maintenance_part_id,
			name,
			quantity,
			COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), '') AS delivery_date
	`

	var part MaintenancePart

	err := r.DB.QueryRow(
		ctx,
		query,
		partID,
		input.Quantity,
		input.DeliveryDate,
	).Scan(
		&part.MaintenancePartID,
		&part.Name,
		&part.Quantity,
		&part.DeliveryDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MaintenancePart{}, ErrMaintenancePartNotFound
		}

		return MaintenancePart{}, err
	}

	return part, nil
}
