package orders

type Order struct {
	OrderNumber int64  `json:"order_number"`
	CarID       int64  `json:"car_id"`
	ServiceID   int64  `json:"service_id"`
	ReadyDate   string `json:"ready_date"`
	CreatedAt   string `json:"created_at"`
}

type OrderView struct {
	OrderNumber    int64  `json:"order_number"`
	CarPlateNumber string `json:"car_plate_number"`
	OwnerFullName  string `json:"owner_full_name"`
	OwnerPhone     string `json:"owner_phone"`
	ServiceName    string `json:"service_name"`
	ReadyDate      string `json:"ready_date"`
	Employees      string `json:"employees"`
}

type OrderEmployee struct {
	EmployeeID int64  `json:"employee_id"`
	FullName   string `json:"full_name"`
	Specialty  string `json:"specialty"`
}

type OrderRepairPart struct {
	RepairPartID int64  `json:"repair_part_id"`
	Name         string `json:"name"`
	QuantityUsed int    `json:"quantity_used"`
}

type OrderMaintenancePart struct {
	MaintenancePartID int64  `json:"maintenance_part_id"`
	Name              string `json:"name"`
	QuantityUsed      int    `json:"quantity_used"`
}

type OrderDetails struct {
	OrderNumber      int64                  `json:"order_number"`
	CarPlateNumber   string                 `json:"car_plate_number"`
	OwnerFullName    string                 `json:"owner_full_name"`
	OwnerPhone       string                 `json:"owner_phone"`
	ServiceName      string                 `json:"service_name"`
	ReadyDate        string                 `json:"ready_date"`
	Employees        []OrderEmployee        `json:"employees"`
	RepairParts      []OrderRepairPart      `json:"repair_parts"`
	MaintenanceParts []OrderMaintenancePart `json:"maintenance_parts"`
}

type CreateOrderInput struct {
	CarID     int64  `json:"car_id"`
	ServiceID int64  `json:"service_id"`
	ReadyDate string `json:"ready_date"`
}

type UpdateOrderInput struct {
	CarID     int64  `json:"car_id"`
	ServiceID int64  `json:"service_id"`
	ReadyDate string `json:"ready_date"`
}

type AssignEmployeesInput struct {
	OrderNumber int64   `json:"order_number"`
	EmployeeIDs []int64 `json:"employee_ids"`
}

type ReplaceEmployeesInput struct {
	EmployeeIDs []int64 `json:"employee_ids"`
}

type RepairPartItemInput struct {
	RepairPartID int64 `json:"repair_part_id"`
	QuantityUsed int   `json:"quantity_used"`
}

type MaintenancePartItemInput struct {
	MaintenancePartID int64 `json:"maintenance_part_id"`
	QuantityUsed      int   `json:"quantity_used"`
}

type AddRepairPartsInput struct {
	OrderNumber int64                 `json:"order_number"`
	Items       []RepairPartItemInput `json:"items"`
}

type ReplaceRepairPartsInput struct {
	Items []RepairPartItemInput `json:"items"`
}

type AddMaintenancePartsInput struct {
	OrderNumber int64                      `json:"order_number"`
	Items       []MaintenancePartItemInput `json:"items"`
}

type ReplaceMaintenancePartsInput struct {
	Items []MaintenancePartItemInput `json:"items"`
}

type OrderFormCarOption struct {
	CarID         int64  `json:"car_id"`
	PlateNumber   string `json:"plate_number"`
	Brand         string `json:"brand"`
	OwnerFullName string `json:"owner_full_name"`
	Label         string `json:"label"`
}

type OrderFormServiceOption struct {
	ServiceID              int64   `json:"service_id"`
	Name                   string  `json:"name"`
	PriceRub               float64 `json:"price_rub"`
	RegularDiscountPercent float64 `json:"regular_discount_percent"`
	DiscountedPriceRub     float64 `json:"discounted_price_rub"`
}

type OrderFormEmployeeOption struct {
	EmployeeID int64  `json:"employee_id"`
	FullName   string `json:"full_name"`
	Specialty  string `json:"specialty"`
}

type OrderFormRepairPartOption struct {
	RepairPartID int64  `json:"repair_part_id"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
}

type OrderFormMaintenancePartOption struct {
	MaintenancePartID int64  `json:"maintenance_part_id"`
	Name              string `json:"name"`
	Quantity          int    `json:"quantity"`
}

type OrderFormData struct {
	Cars             []OrderFormCarOption             `json:"cars"`
	Services         []OrderFormServiceOption         `json:"services"`
	Employees        []OrderFormEmployeeOption        `json:"employees"`
	RepairParts      []OrderFormRepairPartOption      `json:"repair_parts"`
	MaintenanceParts []OrderFormMaintenancePartOption `json:"maintenance_parts"`
}
