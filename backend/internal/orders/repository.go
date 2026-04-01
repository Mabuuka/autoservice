package orders

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOrderNotFound = errors.New("order not found")

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

func (r *Repository) Update(ctx context.Context, orderNumber int64, input UpdateOrderInput) (Order, error) {
	query := `
		UPDATE automaster.orders
		SET
			car_id = $2,
			service_id = $3,
			ready_date = $4
		WHERE order_number = $1
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
		orderNumber,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	return order, nil
}

func (r *Repository) Delete(ctx context.Context, orderNumber int64) error {
	query := `
		DELETE FROM automaster.orders
		WHERE order_number = $1
	`

	result, err := r.DB.Exec(ctx, query, orderNumber)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrOrderNotFound
	}

	return nil
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

func (r *Repository) ReplaceEmployees(ctx context.Context, orderNumber int64, employeeIDs []int64) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureOrderExists(ctx, tx, orderNumber); err != nil {
		return err
	}

	deleteQuery := `
		DELETE FROM automaster.order_employees
		WHERE order_number = $1
	`

	if _, err := tx.Exec(ctx, deleteQuery, orderNumber); err != nil {
		return err
	}

	if len(employeeIDs) > 0 {
		insertQuery := `
			INSERT INTO automaster.order_employees (order_number, employee_id)
			SELECT $1, UNNEST($2::bigint[])
		`

		if _, err := tx.Exec(ctx, insertQuery, orderNumber, employeeIDs); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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

func (r *Repository) ReplaceRepairParts(ctx context.Context, orderNumber int64, items []RepairPartItemInput) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureOrderExists(ctx, tx, orderNumber); err != nil {
		return err
	}

	deleteQuery := `
		DELETE FROM automaster.order_repair_parts
		WHERE order_number = $1
	`

	if _, err := tx.Exec(ctx, deleteQuery, orderNumber); err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO automaster.order_repair_parts (order_number, repair_part_id, quantity_used)
		VALUES ($1, $2, $3)
	`

	for _, item := range items {
		if _, err := tx.Exec(ctx, insertQuery, orderNumber, item.RepairPartID, item.QuantityUsed); err != nil {
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

func (r *Repository) ReplaceMaintenanceParts(ctx context.Context, orderNumber int64, items []MaintenancePartItemInput) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureOrderExists(ctx, tx, orderNumber); err != nil {
		return err
	}

	deleteQuery := `
		DELETE FROM automaster.order_maintenance_parts
		WHERE order_number = $1
	`

	if _, err := tx.Exec(ctx, deleteQuery, orderNumber); err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO automaster.order_maintenance_parts (order_number, maintenance_part_id, quantity_used)
		VALUES ($1, $2, $3)
	`

	for _, item := range items {
		if _, err := tx.Exec(ctx, insertQuery, orderNumber, item.MaintenancePartID, item.QuantityUsed); err != nil {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderDetails{}, ErrOrderNotFound
		}
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

func (r *Repository) GetFormData(ctx context.Context) (OrderFormData, error) {
	cars, err := r.getOrderFormCars(ctx)
	if err != nil {
		return OrderFormData{}, err
	}

	services, err := r.getOrderFormServices(ctx)
	if err != nil {
		return OrderFormData{}, err
	}

	employees, err := r.getOrderFormEmployees(ctx)
	if err != nil {
		return OrderFormData{}, err
	}

	repairParts, err := r.getOrderFormRepairParts(ctx)
	if err != nil {
		return OrderFormData{}, err
	}

	maintenanceParts, err := r.getOrderFormMaintenanceParts(ctx)
	if err != nil {
		return OrderFormData{}, err
	}

	return OrderFormData{
		Cars:             cars,
		Services:         services,
		Employees:        employees,
		RepairParts:      repairParts,
		MaintenanceParts: maintenanceParts,
	}, nil
}

func (r *Repository) getOrderFormCars(ctx context.Context) ([]OrderFormCarOption, error) {
	query := `
		SELECT
			c.car_id,
			c.plate_number,
			c.brand,
			o.full_name,
			c.brand || ' • ' || c.plate_number || ' • ' || o.full_name AS label
		FROM automaster.cars c
		JOIN automaster.owners o ON o.owner_id = c.owner_id
		ORDER BY c.car_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := make([]OrderFormCarOption, 0)
	for rows.Next() {
		var car OrderFormCarOption
		if err := rows.Scan(&car.CarID, &car.PlateNumber, &car.Brand, &car.OwnerFullName, &car.Label); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *Repository) getOrderFormServices(ctx context.Context) ([]OrderFormServiceOption, error) {
	query := `
		SELECT service_id, name, price_rub::float8, regular_discount_percent::float8, discounted_price_rub::float8
		FROM automaster.services
		ORDER BY service_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]OrderFormServiceOption, 0)
	for rows.Next() {
		var service OrderFormServiceOption
		if err := rows.Scan(
			&service.ServiceID,
			&service.Name,
			&service.PriceRub,
			&service.RegularDiscountPercent,
			&service.DiscountedPriceRub,
		); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repository) getOrderFormEmployees(ctx context.Context) ([]OrderFormEmployeeOption, error) {
	query := `
		SELECT employee_id, full_name, specialty
		FROM automaster.employees
		ORDER BY employee_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]OrderFormEmployeeOption, 0)
	for rows.Next() {
		var employee OrderFormEmployeeOption
		if err := rows.Scan(&employee.EmployeeID, &employee.FullName, &employee.Specialty); err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *Repository) getOrderFormRepairParts(ctx context.Context) ([]OrderFormRepairPartOption, error) {
	query := `
		SELECT repair_part_id, name, quantity
		FROM automaster.repair_parts
		ORDER BY repair_part_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repairParts := make([]OrderFormRepairPartOption, 0)
	for rows.Next() {
		var repairPart OrderFormRepairPartOption
		if err := rows.Scan(&repairPart.RepairPartID, &repairPart.Name, &repairPart.Quantity); err != nil {
			return nil, err
		}
		repairParts = append(repairParts, repairPart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return repairParts, nil
}

func (r *Repository) getOrderFormMaintenanceParts(ctx context.Context) ([]OrderFormMaintenancePartOption, error) {
	query := `
		SELECT maintenance_part_id, name, quantity
		FROM automaster.maintenance_parts
		ORDER BY maintenance_part_id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	maintenanceParts := make([]OrderFormMaintenancePartOption, 0)
	for rows.Next() {
		var maintenancePart OrderFormMaintenancePartOption
		if err := rows.Scan(&maintenancePart.MaintenancePartID, &maintenancePart.Name, &maintenancePart.Quantity); err != nil {
			return nil, err
		}
		maintenanceParts = append(maintenanceParts, maintenancePart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return maintenanceParts, nil
}

func ensureOrderExists(ctx context.Context, tx pgx.Tx, orderNumber int64) error {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM automaster.orders
			WHERE order_number = $1
		)
	`

	var exists bool
	if err := tx.QueryRow(ctx, query, orderNumber).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		return ErrOrderNotFound
	}

	return nil
}
