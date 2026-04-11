package repairparts

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

	parts, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch repair parts.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, parts)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateRepairPartInput

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
			api.Conflict(w, "Repair part with the same name already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "Quantity must be 0 or greater.")
		default:
			api.Internal(w, "Failed to create repair part.")
		}
		return
	}

	_ = api.WriteCreated(w, "Repair part created successfully.", part)
}

func (h *Handler) Restock(w http.ResponseWriter, r *http.Request) {
	partID, err := parseRepairPartID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input RestockRepairPartInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.DeliveryDate = strings.TrimSpace(input.DeliveryDate)

	if input.Quantity <= 0 {
		api.BadRequest(w, "quantity must be greater than 0.")
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

	part, err := h.Repo.Restock(ctx, partID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrRepairPartNotFound):
			api.NotFound(w, "Repair part not found.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "quantity must be greater than 0.")
		default:
			api.Internal(w, "Failed to restock repair part.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Repair part restocked successfully.", part)
}

func parseRepairPartID(r *http.Request) (int64, error) {
	partIDStr := strings.TrimSpace(r.PathValue("id"))
	if partIDStr == "" {
		return 0, errors.New("repair part id is required")
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil || partID <= 0 {
		return 0, errors.New("repair part id must be a positive integer")
	}

	return partID, nil
}
