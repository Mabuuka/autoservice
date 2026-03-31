package employees

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
