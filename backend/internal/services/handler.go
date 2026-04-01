package services

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

	services, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch services.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, services)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateServiceInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" || input.PriceRub < 0 || input.RegularDiscountPercent < 0 || input.RegularDiscountPercent > 100 {
		api.BadRequest(w, "Field name is required, price_rub must be >= 0, and regular_discount_percent must be between 0 and 100.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	service, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Service with the same name already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "Service price or discount value is outside the allowed range.")
		default:
			api.Internal(w, "Failed to create service.")
		}
		return
	}

	_ = api.WriteCreated(w, "Service created successfully.", service)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	serviceID, err := parseServiceID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input UpdateServiceInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" || input.PriceRub < 0 || input.RegularDiscountPercent < 0 || input.RegularDiscountPercent > 100 {
		api.BadRequest(w, "Field name is required, price_rub must be >= 0, and regular_discount_percent must be between 0 and 100.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	service, err := h.Repo.Update(ctx, serviceID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrServiceNotFound):
			api.NotFound(w, "Service not found.")
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Service with the same name already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "Service price or discount value is outside the allowed range.")
		default:
			api.Internal(w, "Failed to update service.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Service updated successfully.", service)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	serviceID, err := parseServiceID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Delete(ctx, serviceID); err != nil {
		switch {
		case errors.Is(err, ErrServiceNotFound):
			api.NotFound(w, "Service not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Cannot delete service because it is linked to existing orders.")
		default:
			api.Internal(w, "Failed to delete service.")
		}
		return
	}

	_ = api.WriteDeleted(w, "Service deleted successfully.", serviceID)
}

func parseServiceID(r *http.Request) (int64, error) {
	serviceIDStr := strings.TrimSpace(r.PathValue("id"))
	if serviceIDStr == "" {
		return 0, errors.New("service id is required")
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil || serviceID <= 0 {
		return 0, errors.New("service id must be a positive integer")
	}

	return serviceID, nil
}
