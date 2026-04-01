package maintenanceparts

import (
	"context"
	"net/http"
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

	parts, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch maintenance parts.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, parts)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateMaintenancePartInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.DeliveryDate = strings.TrimSpace(input.DeliveryDate)

	if input.Name == "" || input.Quantity < 0 {
		api.BadRequest(w, "Field name is required and quantity must be 0 or greater.")
		return
	}

	if input.DeliveryDate != "" {
		if _, err := time.Parse("2006-01-02", input.DeliveryDate); err != nil {
			api.BadRequest(w, "delivery_date must be in YYYY-MM-DD format.")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	part, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Maintenance part with the same name already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "Quantity must be 0 or greater.")
		default:
			api.Internal(w, "Failed to create maintenance part.")
		}
		return
	}

	_ = api.WriteCreated(w, "Maintenance part created successfully.", part)
}
