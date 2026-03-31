package owners

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

func (r *Repository) GetAll(ctx context.Context) ([]Owner, error) {
	query := `
		SELECT owner_id, full_name, phone, driver_license, is_regular
		FROM automaster.owners
		ORDER BY owner_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	owners := make([]Owner, 0)

	for rows.Next() {
		var owner Owner

		err := rows.Scan(
			&owner.OwnerID,
			&owner.FullName,
			&owner.Phone,
			&owner.DriverLicense,
			&owner.IsRegular,
		)
		if err != nil {
			return nil, err
		}

		owners = append(owners, owner)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return owners, nil
}

func (r *Repository) Create(ctx context.Context, input CreateOwnerInput) (Owner, error) {
	query := `
		INSERT INTO automaster.owners (full_name, phone, driver_license, is_regular)
		VALUES ($1, $2, $3, $4)
		RETURNING owner_id, full_name, phone, driver_license, is_regular
	`

	var owner Owner

	err := r.DB.QueryRow(
		ctx,
		query,
		input.FullName,
		input.Phone,
		input.DriverLicense,
		input.IsRegular,
	).Scan(
		&owner.OwnerID,
		&owner.FullName,
		&owner.Phone,
		&owner.DriverLicense,
		&owner.IsRegular,
	)
	if err != nil {
		return Owner{}, err
	}

	return owner, nil
}
