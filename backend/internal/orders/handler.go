package orders

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"autoservice/backend/internal/api"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	orders, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch orders.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, orders)
}

func (h *Handler) GetFormData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	formData, err := h.Repo.GetFormData(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch order form data.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, formData)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateOrderInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.ReadyDate = strings.TrimSpace(input.ReadyDate)

	if input.CarID <= 0 || input.ServiceID <= 0 || input.ReadyDate == "" {
		api.BadRequest(w, "Fields car_id, service_id and ready_date are required.")
		return
	}

	if _, err := time.Parse("2006-01-02", input.ReadyDate); err != nil {
		api.BadRequest(w, "ready_date must be in YYYY-MM-DD format.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	order, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Car or service with the specified id does not exist.")
		default:
			api.Internal(w, "Failed to create order.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, order.OrderNumber)
	if err != nil {
		api.Internal(w, "Failed to fetch created order details.")
		return
	}

	_ = api.WriteCreated(w, "Order created successfully.", details)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input UpdateOrderInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.ReadyDate = strings.TrimSpace(input.ReadyDate)

	if input.CarID <= 0 || input.ServiceID <= 0 || input.ReadyDate == "" {
		api.BadRequest(w, "Fields car_id, service_id and ready_date are required.")
		return
	}

	if _, err := time.Parse("2006-01-02", input.ReadyDate); err != nil {
		api.BadRequest(w, "ready_date must be in YYYY-MM-DD format.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, err = h.Repo.Update(ctx, orderNumber, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			api.NotFound(w, "Order not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Car or service with the specified id does not exist.")
		default:
			api.Internal(w, "Failed to update order.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Order updated successfully.", details)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Delete(ctx, orderNumber); err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			api.NotFound(w, "Order not found.")
		default:
			api.Internal(w, "Failed to delete order.")
		}
		return
	}

	_ = api.WriteDeleted(w, "Order deleted successfully.", orderNumber)
}

func (h *Handler) AssignEmployees(w http.ResponseWriter, r *http.Request) {
	var input AssignEmployeesInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	if input.OrderNumber <= 0 || len(input.EmployeeIDs) == 0 {
		api.BadRequest(w, "Fields order_number and employee_ids are required.")
		return
	}

	normalizedEmployeeIDs, err := normalizeEmployeeIDs(input.EmployeeIDs)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	input.EmployeeIDs = normalizedEmployeeIDs

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AssignEmployees(ctx, input); err != nil {
		switch {
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Order or employee with the specified id does not exist.")
		default:
			api.Internal(w, "Failed to assign employees to order.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Employees assigned successfully.", details)
}

func (h *Handler) ReplaceEmployees(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input ReplaceEmployeesInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.EmployeeIDs, err = normalizeEmployeeIDs(input.EmployeeIDs)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.ReplaceEmployees(ctx, orderNumber, input.EmployeeIDs); err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			api.NotFound(w, "Order not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "One or more employees with the specified ids do not exist.")
		default:
			api.Internal(w, "Failed to replace order employees.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Order employees updated successfully.", details)
}

func (h *Handler) AddRepairParts(w http.ResponseWriter, r *http.Request) {
	var input AddRepairPartsInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	if input.OrderNumber <= 0 || len(input.Items) == 0 {
		api.BadRequest(w, "Fields order_number and items are required.")
		return
	}

	normalizedItems, err := normalizeRepairPartItems(input.Items)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	input.Items = normalizedItems

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AddRepairParts(ctx, input); err != nil {
		switch {
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Order or repair part with the specified id does not exist.")
		default:
			api.Internal(w, "Failed to add repair parts to order.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Repair parts added successfully.", details)
}

func (h *Handler) ReplaceRepairParts(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input ReplaceRepairPartsInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Items, err = normalizeRepairPartItemsAllowEmpty(input.Items)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.ReplaceRepairParts(ctx, orderNumber, input.Items); err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			api.NotFound(w, "Order not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "One or more repair parts with the specified ids do not exist.")
		default:
			api.Internal(w, "Failed to replace order repair parts.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Order repair parts updated successfully.", details)
}

func (h *Handler) AddMaintenanceParts(w http.ResponseWriter, r *http.Request) {
	var input AddMaintenancePartsInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	if input.OrderNumber <= 0 || len(input.Items) == 0 {
		api.BadRequest(w, "Fields order_number and items are required.")
		return
	}

	normalizedItems, err := normalizeMaintenancePartItems(input.Items)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	input.Items = normalizedItems

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AddMaintenanceParts(ctx, input); err != nil {
		switch {
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Order or maintenance part with the specified id does not exist.")
		default:
			api.Internal(w, "Failed to add maintenance parts to order.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Maintenance parts added successfully.", details)
}

func (h *Handler) ReplaceMaintenanceParts(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input ReplaceMaintenancePartsInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Items, err = normalizeMaintenancePartItemsAllowEmpty(input.Items)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.ReplaceMaintenanceParts(ctx, orderNumber, input.Items); err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			api.NotFound(w, "Order not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "One or more maintenance parts with the specified ids do not exist.")
		default:
			api.Internal(w, "Failed to replace order maintenance parts.")
		}
		return
	}

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch updated order details.")
		return
	}

	_ = api.WriteUpdated(w, "Order maintenance parts updated successfully.", details)
}

func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromQuery(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	h.writeOrderDetails(w, r, orderNumber)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	orderNumber, err := parseOrderNumberFromPath(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	h.writeOrderDetails(w, r, orderNumber)
}

func (h *Handler) writeOrderDetails(w http.ResponseWriter, r *http.Request, orderNumber int64) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			api.NotFound(w, "Order not found.")
			return
		}
		api.Internal(w, "Failed to fetch order details.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, details)
}

func parseOrderNumberFromQuery(r *http.Request) (int64, error) {
	orderNumberStr := strings.TrimSpace(r.URL.Query().Get("order_number"))
	if orderNumberStr == "" {
		return 0, errors.New("query parameter order_number is required")
	}

	return parsePositiveInt64(orderNumberStr, "order_number must be a positive integer")
}

func parseOrderNumberFromPath(r *http.Request) (int64, error) {
	orderNumberStr := strings.TrimSpace(r.PathValue("id"))
	if orderNumberStr == "" {
		return 0, errors.New("order id is required")
	}

	return parsePositiveInt64(orderNumberStr, "order id must be a positive integer")
}

func parsePositiveInt64(value string, message string) (int64, error) {
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsedValue <= 0 {
		return 0, errors.New(message)
	}

	return parsedValue, nil
}

func normalizeEmployeeIDs(employeeIDs []int64) ([]int64, error) {
	if employeeIDs == nil {
		return []int64{}, nil
	}

	unique := make(map[int64]struct{})
	result := make([]int64, 0, len(employeeIDs))

	for _, employeeID := range employeeIDs {
		if employeeID <= 0 {
			return nil, errors.New("employee_ids must contain only positive integers")
		}
		if _, exists := unique[employeeID]; exists {
			continue
		}
		unique[employeeID] = struct{}{}
		result = append(result, employeeID)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

func normalizeRepairPartItems(items []RepairPartItemInput) ([]RepairPartItemInput, error) {
	if len(items) == 0 {
		return nil, errors.New("items are required")
	}
	return normalizeRepairPartItemsAllowEmpty(items)
}

func normalizeRepairPartItemsAllowEmpty(items []RepairPartItemInput) ([]RepairPartItemInput, error) {
	if items == nil {
		return []RepairPartItemInput{}, nil
	}

	aggregated := make(map[int64]int)
	for _, item := range items {
		if item.RepairPartID <= 0 || item.QuantityUsed <= 0 {
			return nil, errors.New("repair_part_id and quantity_used must be positive integers")
		}
		aggregated[item.RepairPartID] += item.QuantityUsed
	}

	partIDs := make([]int64, 0, len(aggregated))
	for repairPartID := range aggregated {
		partIDs = append(partIDs, repairPartID)
	}
	sort.Slice(partIDs, func(i, j int) bool {
		return partIDs[i] < partIDs[j]
	})

	result := make([]RepairPartItemInput, 0, len(partIDs))
	for _, repairPartID := range partIDs {
		result = append(result, RepairPartItemInput{
			RepairPartID: repairPartID,
			QuantityUsed: aggregated[repairPartID],
		})
	}

	return result, nil
}

func normalizeMaintenancePartItems(items []MaintenancePartItemInput) ([]MaintenancePartItemInput, error) {
	if len(items) == 0 {
		return nil, errors.New("items are required")
	}
	return normalizeMaintenancePartItemsAllowEmpty(items)
}

func normalizeMaintenancePartItemsAllowEmpty(items []MaintenancePartItemInput) ([]MaintenancePartItemInput, error) {
	if items == nil {
		return []MaintenancePartItemInput{}, nil
	}

	aggregated := make(map[int64]int)
	for _, item := range items {
		if item.MaintenancePartID <= 0 || item.QuantityUsed <= 0 {
			return nil, errors.New("maintenance_part_id and quantity_used must be positive integers")
		}
		aggregated[item.MaintenancePartID] += item.QuantityUsed
	}

	partIDs := make([]int64, 0, len(aggregated))
	for maintenancePartID := range aggregated {
		partIDs = append(partIDs, maintenancePartID)
	}
	sort.Slice(partIDs, func(i, j int) bool {
		return partIDs[i] < partIDs[j]
	})

	result := make([]MaintenancePartItemInput, 0, len(partIDs))
	for _, maintenancePartID := range partIDs {
		result = append(result, MaintenancePartItemInput{
			MaintenancePartID: maintenancePartID,
			QuantityUsed:      aggregated[maintenancePartID],
		})
	}

	return result, nil
}
