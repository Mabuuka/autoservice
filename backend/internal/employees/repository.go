package employees

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmployeeNotFound = errors.New("employee not found")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]Employee, error) {
	query := `
		SELECT employee_id, COALESCE(personnel_number, 0)::int, specialty, phone, full_name
		FROM automaster.employees
		ORDER BY employee_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]Employee, 0)

	for rows.Next() {
		var employee Employee

		err := rows.Scan(
			&employee.EmployeeID,
			&employee.PersonnelNumber,
			&employee.Specialty,
			&employee.Phone,
			&employee.FullName,
		)
		if err != nil {
			return nil, err
		}

		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *Repository) Create(ctx context.Context, input CreateEmployeeInput) (Employee, error) {
	query := `
		INSERT INTO automaster.employees (personnel_number, specialty, phone, full_name)
		VALUES ($1, $2, $3, $4)
		RETURNING employee_id, COALESCE(personnel_number, 0)::int, specialty, phone, full_name
	`

	var employee Employee

	err := r.DB.QueryRow(
		ctx,
		query,
		input.PersonnelNumber,
		input.Specialty,
		input.Phone,
		input.FullName,
	).Scan(
		&employee.EmployeeID,
		&employee.PersonnelNumber,
		&employee.Specialty,
		&employee.Phone,
		&employee.FullName,
	)
	if err != nil {
		return Employee{}, err
	}

	return employee, nil
}

func (r *Repository) Update(ctx context.Context, employeeID int64, input UpdateEmployeeInput) (Employee, error) {
	query := `
		UPDATE automaster.employees
		SET
			personnel_number = $2,
			specialty = $3,
			phone = $4,
			full_name = $5
		WHERE employee_id = $1
		RETURNING employee_id, COALESCE(personnel_number, 0)::int, specialty, phone, full_name
	`

	var employee Employee

	err := r.DB.QueryRow(
		ctx,
		query,
		employeeID,
		input.PersonnelNumber,
		input.Specialty,
		input.Phone,
		input.FullName,
	).Scan(
		&employee.EmployeeID,
		&employee.PersonnelNumber,
		&employee.Specialty,
		&employee.Phone,
		&employee.FullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Employee{}, ErrEmployeeNotFound
		}

		return Employee{}, err
	}

	return employee, nil
}

func (r *Repository) Delete(ctx context.Context, employeeID int64) error {
	query := `
		DELETE FROM automaster.employees
		WHERE employee_id = $1
	`

	result, err := r.DB.Exec(ctx, query, employeeID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}
