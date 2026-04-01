package owners

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

	owners, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch owners.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, owners)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateOwnerInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.DriverLicense = strings.TrimSpace(input.DriverLicense)

	if input.FullName == "" || input.Phone == "" || input.DriverLicense == "" {
		api.BadRequest(w, "Fields full_name, phone and driver_license are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	owner, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Owner with the same full_name or phone already exists.")
		default:
			api.Internal(w, "Failed to create owner.")
		}
		return
	}

	_ = api.WriteCreated(w, "Owner created successfully.", owner)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ownerID, err := parseOwnerID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input UpdateOwnerInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.DriverLicense = strings.TrimSpace(input.DriverLicense)

	if input.FullName == "" || input.Phone == "" || input.DriverLicense == "" {
		api.BadRequest(w, "Fields full_name, phone and driver_license are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	owner, err := h.Repo.Update(ctx, ownerID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrOwnerNotFound):
			api.NotFound(w, "Owner not found.")
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Owner with the same full_name or phone already exists.")
		default:
			api.Internal(w, "Failed to update owner.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Owner updated successfully.", owner)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ownerID, err := parseOwnerID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Delete(ctx, ownerID); err != nil {
		switch {
		case errors.Is(err, ErrOwnerNotFound):
			api.NotFound(w, "Owner not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Cannot delete owner because it is linked to existing cars.")
		default:
			api.Internal(w, "Failed to delete owner.")
		}
		return
	}

	_ = api.WriteDeleted(w, "Owner deleted successfully.", ownerID)
}

func parseOwnerID(r *http.Request) (int64, error) {
	ownerIDStr := strings.TrimSpace(r.PathValue("id"))
	if ownerIDStr == "" {
		return 0, errors.New("owner id is required")
	}

	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil || ownerID <= 0 {
		return 0, errors.New("owner id must be a positive integer")
	}

	return ownerID, nil
}
