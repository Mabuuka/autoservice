package orders

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]OrderView, error) {
	query := `
		SELECT
			order_number,
			car_plate_number,
			owner_full_name,
			owner_phone,
			service_name,
			TO_CHAR(ready_date, 'YYYY-MM-DD') AS ready_date,
			COALESCE(employees, '') AS employees
		FROM automaster.v_workshop_orders
		ORDER BY order_number
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]OrderView, 0)

	for rows.Next() {
		var order OrderView

		err := rows.Scan(
			&order.OrderNumber,
			&order.CarPlateNumber,
			&order.OwnerFullName,
			&order.OwnerPhone,
			&order.ServiceName,
			&order.ReadyDate,
			&order.Employees,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) Create(ctx context.Context, input CreateOrderInput) (Order, error) {
	query := `
		INSERT INTO automaster.orders (car_id, service_id, ready_date)
		VALUES ($1, $2, $3)
		RETURNING
			order_number,
			car_id,
			service_id,
			TO_CHAR(ready_date, 'YYYY-MM-DD') AS ready_date,
			TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS') AS created_at
	`

	var order Order

	err := r.DB.QueryRow(
		ctx,
		query,
		input.CarID,
		input.ServiceID,
		input.ReadyDate,
	).Scan(
		&order.OrderNumber,
		&order.CarID,
		&order.ServiceID,
		&order.ReadyDate,
		&order.CreatedAt,
	)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

func (r *Repository) AssignEmployees(ctx context.Context, input AssignEmployeesInput) error {
	if len(input.EmployeeIDs) == 0 {
		return errors.New("employee_ids are required")
	}

	query := `
		INSERT INTO automaster.order_employees (order_number, employee_id)
		SELECT $1, UNNEST($2::bigint[])
		ON CONFLICT (order_number, employee_id) DO NOTHING
	`

	_, err := r.DB.Exec(ctx, query, input.OrderNumber, input.EmployeeIDs)
	return err
}

func (r *Repository) AddRepairParts(ctx context.Context, input AddRepairPartsInput) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO automaster.order_repair_parts (order_number, repair_part_id, quantity_used)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_number, repair_part_id)
		DO UPDATE SET quantity_used = automaster.order_repair_parts.quantity_used + EXCLUDED.quantity_used
	`

	for _, item := range input.Items {
		if _, err := tx.Exec(ctx, query, input.OrderNumber, item.RepairPartID, item.QuantityUsed); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) AddMaintenanceParts(ctx context.Context, input AddMaintenancePartsInput) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO automaster.order_maintenance_parts (order_number, maintenance_part_id, quantity_used)
		VALUES ($1, $2, $3)
		ON CONFLICT (order_number, maintenance_part_id)
		DO UPDATE SET quantity_used = automaster.order_maintenance_parts.quantity_used + EXCLUDED.quantity_used
	`

	for _, item := range input.Items {
		if _, err := tx.Exec(ctx, query, input.OrderNumber, item.MaintenancePartID, item.QuantityUsed); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetDetails(ctx context.Context, orderNumber int64) (OrderDetails, error) {
	baseQuery := `
		SELECT
			order_number,
			car_plate_number,
			owner_full_name,
			owner_phone,
			service_name,
			TO_CHAR(ready_date, 'YYYY-MM-DD') AS ready_date
		FROM automaster.v_workshop_orders
		WHERE order_number = $1
	`

	var details OrderDetails
	if err := r.DB.QueryRow(ctx, baseQuery, orderNumber).Scan(
		&details.OrderNumber,
		&details.CarPlateNumber,
		&details.OwnerFullName,
		&details.OwnerPhone,
		&details.ServiceName,
		&details.ReadyDate,
	); err != nil {
		return OrderDetails{}, err
	}

	employeesQuery := `
		SELECT e.employee_id, e.full_name, e.specialty
		FROM automaster.order_employees oe
		JOIN automaster.employees e ON e.employee_id = oe.employee_id
		WHERE oe.order_number = $1
		ORDER BY e.employee_id
	`

	employeeRows, err := r.DB.Query(ctx, employeesQuery, orderNumber)
	if err != nil {
		return OrderDetails{}, err
	}
	defer employeeRows.Close()

	details.Employees = make([]OrderEmployee, 0)
	for employeeRows.Next() {
		var employee OrderEmployee
		if err := employeeRows.Scan(&employee.EmployeeID, &employee.FullName, &employee.Specialty); err != nil {
			return OrderDetails{}, err
		}
		details.Employees = append(details.Employees, employee)
	}
	if err := employeeRows.Err(); err != nil {
		return OrderDetails{}, err
	}

	repairPartsQuery := `
		SELECT rp.repair_part_id, rp.name, orp.quantity_used
		FROM automaster.order_repair_parts orp
		JOIN automaster.repair_parts rp ON rp.repair_part_id = orp.repair_part_id
		WHERE orp.order_number = $1
		ORDER BY rp.repair_part_id
	`

	repairRows, err := r.DB.Query(ctx, repairPartsQuery, orderNumber)
	if err != nil {
		return OrderDetails{}, err
	}
	defer repairRows.Close()

	details.RepairParts = make([]OrderRepairPart, 0)
	for repairRows.Next() {
		var item OrderRepairPart
		if err := repairRows.Scan(&item.RepairPartID, &item.Name, &item.QuantityUsed); err != nil {
			return OrderDetails{}, err
		}
		details.RepairParts = append(details.RepairParts, item)
	}
	if err := repairRows.Err(); err != nil {
		return OrderDetails{}, err
	}

	maintenancePartsQuery := `
		SELECT mp.maintenance_part_id, mp.name, omp.quantity_used
		FROM automaster.order_maintenance_parts omp
		JOIN automaster.maintenance_parts mp ON mp.maintenance_part_id = omp.maintenance_part_id
		WHERE omp.order_number = $1
		ORDER BY mp.maintenance_part_id
	`

	maintenanceRows, err := r.DB.Query(ctx, maintenancePartsQuery, orderNumber)
	if err != nil {
		return OrderDetails{}, err
	}
	defer maintenanceRows.Close()

	details.MaintenanceParts = make([]OrderMaintenancePart, 0)
	for maintenanceRows.Next() {
		var item OrderMaintenancePart
		if err := maintenanceRows.Scan(&item.MaintenancePartID, &item.Name, &item.QuantityUsed); err != nil {
			return OrderDetails{}, err
		}
		details.MaintenanceParts = append(details.MaintenanceParts, item)
	}
	if err := maintenanceRows.Err(); err != nil {
		return OrderDetails{}, err
	}

	return details, nil
}
