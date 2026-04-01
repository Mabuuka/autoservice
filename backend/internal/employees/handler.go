package employees

import (
	"context"
	"errors"
	"net/http"
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

	employees, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch employees.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, employees)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateEmployeeInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Specialty = strings.TrimSpace(input.Specialty)
	input.Phone = strings.TrimSpace(input.Phone)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.PersonnelNumber <= 0 || input.Specialty == "" || input.Phone == "" || input.FullName == "" {
		api.BadRequest(w, "Fields personnel_number, specialty, phone and full_name are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	employee, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Employee with the same personnel_number, phone or full_name already exists.")
		default:
			api.Internal(w, "Failed to create employee.")
		}
		return
	}

	_ = api.WriteCreated(w, "Employee created successfully.", employee)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseEmployeeID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input UpdateEmployeeInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Specialty = strings.TrimSpace(input.Specialty)
	input.Phone = strings.TrimSpace(input.Phone)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.PersonnelNumber <= 0 || input.Specialty == "" || input.Phone == "" || input.FullName == "" {
		api.BadRequest(w, "Fields personnel_number, specialty, phone and full_name are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	employee, err := h.Repo.Update(ctx, employeeID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmployeeNotFound):
			api.NotFound(w, "Employee not found.")
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Employee with the same personnel_number, phone or full_name already exists.")
		default:
			api.Internal(w, "Failed to update employee.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Employee updated successfully.", employee)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseEmployeeID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Delete(ctx, employeeID); err != nil {
		switch {
		case errors.Is(err, ErrEmployeeNotFound):
			api.NotFound(w, "Employee not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Cannot delete employee because it is linked to existing orders.")
		default:
			api.Internal(w, "Failed to delete employee.")
		}
		return
	}

	_ = api.WriteDeleted(w, "Employee deleted successfully.", employeeID)
}

func parseEmployeeID(r *http.Request) (int64, error) {
	employeeIDStr := strings.TrimSpace(r.PathValue("id"))
	if employeeIDStr == "" {
		return 0, errors.New("employee id is required")
	}

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil || employeeID <= 0 {
		return 0, errors.New("employee id must be a positive integer")
	}

	return employeeID, nil
}
