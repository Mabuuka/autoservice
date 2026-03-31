package orders

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
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
		http.Error(w, "failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateOrderInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.ReadyDate = strings.TrimSpace(input.ReadyDate)

	if input.CarID <= 0 || input.ServiceID <= 0 || input.ReadyDate == "" {
		http.Error(w, "car_id, service_id and ready_date are required", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", input.ReadyDate); err != nil {
		http.Error(w, "ready_date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	order, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AssignEmployees(w http.ResponseWriter, r *http.Request) {
	var input AssignEmployeesInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if input.OrderNumber <= 0 || len(input.EmployeeIDs) == 0 {
		http.Error(w, "order_number and employee_ids are required", http.StatusBadRequest)
		return
	}

	for _, employeeID := range input.EmployeeIDs {
		if employeeID <= 0 {
			http.Error(w, "employee_ids must contain only positive numbers", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AssignEmployees(ctx, input); err != nil {
		http.Error(w, "failed to assign employees", http.StatusInternalServerError)
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		http.Error(w, "failed to fetch updated order details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddRepairParts(w http.ResponseWriter, r *http.Request) {
	var input AddRepairPartsInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if input.OrderNumber <= 0 || len(input.Items) == 0 {
		http.Error(w, "order_number and items are required", http.StatusBadRequest)
		return
	}

	for _, item := range input.Items {
		if item.RepairPartID <= 0 || item.QuantityUsed <= 0 {
			http.Error(w, "repair_part_id and quantity_used must be positive", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AddRepairParts(ctx, input); err != nil {
		http.Error(w, "failed to add repair parts", http.StatusInternalServerError)
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		http.Error(w, "failed to fetch updated order details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddMaintenanceParts(w http.ResponseWriter, r *http.Request) {
	var input AddMaintenancePartsInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if input.OrderNumber <= 0 || len(input.Items) == 0 {
		http.Error(w, "order_number and items are required", http.StatusBadRequest)
		return
	}

	for _, item := range input.Items {
		if item.MaintenancePartID <= 0 || item.QuantityUsed <= 0 {
			http.Error(w, "maintenance_part_id and quantity_used must be positive", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.AddMaintenanceParts(ctx, input); err != nil {
		http.Error(w, "failed to add maintenance parts", http.StatusInternalServerError)
		return
	}

	details, err := h.Repo.GetDetails(ctx, input.OrderNumber)
	if err != nil {
		http.Error(w, "failed to fetch updated order details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	orderNumberStr := strings.TrimSpace(r.URL.Query().Get("order_number"))
	if orderNumberStr == "" {
		http.Error(w, "order_number query parameter is required", http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.ParseInt(orderNumberStr, 10, 64)
	if err != nil || orderNumber <= 0 {
		http.Error(w, "order_number must be a positive integer", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	details, err := h.Repo.GetDetails(ctx, orderNumber)
	if err != nil {
		http.Error(w, "failed to fetch order details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(details); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
