package repairparts

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]RepairPart, error) {
	query := `
		SELECT
			repair_part_id,
			name,
			quantity,
			COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), '') AS delivery_date
		FROM automaster.repair_parts
		ORDER BY repair_part_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parts := make([]RepairPart, 0)

	for rows.Next() {
		var part RepairPart

		err := rows.Scan(
			&part.RepairPartID,
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

func (r *Repository) Create(ctx context.Context, input CreateRepairPartInput) (RepairPart, error) {
	query := `
		INSERT INTO automaster.repair_parts (name, quantity, delivery_date)
		VALUES ($1, $2, NULLIF($3, '')::date)
		RETURNING
			repair_part_id,
			name,
			quantity,
			COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), '') AS delivery_date
	`

	var part RepairPart

	err := r.DB.QueryRow(
		ctx,
		query,
		input.Name,
		input.Quantity,
		input.DeliveryDate,
	).Scan(
		&part.RepairPartID,
		&part.Name,
		&part.Quantity,
		&part.DeliveryDate,
	)
	if err != nil {
		return RepairPart{}, err
	}

	return part, nil
}
