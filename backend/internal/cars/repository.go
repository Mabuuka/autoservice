package cars

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCarNotFound = errors.New("car not found")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]Car, error) {
	query := `
		SELECT c.car_id, c.owner_id, o.full_name, c.brand, c.plate_number, c.manufacture_year, c.color
		FROM automaster.cars c
		JOIN automaster.owners o ON o.owner_id = c.owner_id
		ORDER BY c.car_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := make([]Car, 0)

	for rows.Next() {
		var car Car

		err := rows.Scan(
			&car.CarID,
			&car.OwnerID,
			&car.OwnerFullName,
			&car.Brand,
			&car.PlateNumber,
			&car.ManufactureYear,
			&car.Color,
		)
		if err != nil {
			return nil, err
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *Repository) Create(ctx context.Context, input CreateCarInput) (Car, error) {
	query := `
		WITH inserted AS (
			INSERT INTO automaster.cars (owner_id, brand, plate_number, manufacture_year, color)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING car_id, owner_id, brand, plate_number, manufacture_year, color
		)
		SELECT i.car_id, i.owner_id, o.full_name, i.brand, i.plate_number, i.manufacture_year, i.color
		FROM inserted i
		JOIN automaster.owners o ON o.owner_id = i.owner_id
	`

	var car Car

	err := r.DB.QueryRow(
		ctx,
		query,
		input.OwnerID,
		input.Brand,
		input.PlateNumber,
		input.ManufactureYear,
		input.Color,
	).Scan(
		&car.CarID,
		&car.OwnerID,
		&car.OwnerFullName,
		&car.Brand,
		&car.PlateNumber,
		&car.ManufactureYear,
		&car.Color,
	)
	if err != nil {
		return Car{}, err
	}

	return car, nil
}

func (r *Repository) Update(ctx context.Context, carID int64, input UpdateCarInput) (Car, error) {
	query := `
		WITH updated AS (
			UPDATE automaster.cars
			SET
				owner_id = $2,
				brand = $3,
				plate_number = $4,
				manufacture_year = $5,
				color = $6
			WHERE car_id = $1
			RETURNING car_id, owner_id, brand, plate_number, manufacture_year, color
		)
		SELECT u.car_id, u.owner_id, o.full_name, u.brand, u.plate_number, u.manufacture_year, u.color
		FROM updated u
		JOIN automaster.owners o ON o.owner_id = u.owner_id
	`

	var car Car

	err := r.DB.QueryRow(
		ctx,
		query,
		carID,
		input.OwnerID,
		input.Brand,
		input.PlateNumber,
		input.ManufactureYear,
		input.Color,
	).Scan(
		&car.CarID,
		&car.OwnerID,
		&car.OwnerFullName,
		&car.Brand,
		&car.PlateNumber,
		&car.ManufactureYear,
		&car.Color,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Car{}, ErrCarNotFound
		}

		return Car{}, err
	}

	return car, nil
}

func (r *Repository) Delete(ctx context.Context, carID int64) error {
	query := `
		DELETE FROM automaster.cars
		WHERE car_id = $1
	`

	result, err := r.DB.Exec(ctx, query, carID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCarNotFound
	}

	return nil
}
