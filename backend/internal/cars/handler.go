package cars

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

	cars, err := h.Repo.GetAll(ctx)
	if err != nil {
		api.Internal(w, "Failed to fetch cars.")
		return
	}

	_ = api.WriteData(w, http.StatusOK, cars)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateCarInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Brand = strings.TrimSpace(input.Brand)
	input.PlateNumber = strings.TrimSpace(input.PlateNumber)
	input.Color = strings.TrimSpace(input.Color)

	if input.OwnerID <= 0 || input.Brand == "" || input.PlateNumber == "" || input.ManufactureYear <= 0 || input.Color == "" {
		api.BadRequest(w, "Fields owner_id, brand, plate_number, manufacture_year and color are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	car, err := h.Repo.Create(ctx, input)
	if err != nil {
		switch {
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Owner with the specified owner_id does not exist.")
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Car with the same plate_number already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "manufacture_year is outside the allowed range.")
		default:
			api.Internal(w, "Failed to create car.")
		}
		return
	}

	_ = api.WriteCreated(w, "Car created successfully.", car)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	carID, err := parseCarID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var input UpdateCarInput

	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	input.Brand = strings.TrimSpace(input.Brand)
	input.PlateNumber = strings.TrimSpace(input.PlateNumber)
	input.Color = strings.TrimSpace(input.Color)

	if input.OwnerID <= 0 || input.Brand == "" || input.PlateNumber == "" || input.ManufactureYear <= 0 || input.Color == "" {
		api.BadRequest(w, "Fields owner_id, brand, plate_number, manufacture_year and color are required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	car, err := h.Repo.Update(ctx, carID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrCarNotFound):
			api.NotFound(w, "Car not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Owner with the specified owner_id does not exist.")
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Car with the same plate_number already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "manufacture_year is outside the allowed range.")
		default:
			api.Internal(w, "Failed to update car.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Car updated successfully.", car)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	carID, err := parseCarID(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.Repo.Delete(ctx, carID); err != nil {
		switch {
		case errors.Is(err, ErrCarNotFound):
			api.NotFound(w, "Car not found.")
		case api.IsForeignKeyViolation(err):
			api.Conflict(w, "Cannot delete car because it is linked to existing orders.")
		default:
			api.Internal(w, "Failed to delete car.")
		}
		return
	}

	_ = api.WriteDeleted(w, "Car deleted successfully.", carID)
}

func parseCarID(r *http.Request) (int64, error) {
	carIDStr := strings.TrimSpace(r.PathValue("id"))
	if carIDStr == "" {
		return 0, errors.New("car id is required")
	}

	carID, err := strconv.ParseInt(carIDStr, 10, 64)
	if err != nil || carID <= 0 {
		return 0, errors.New("car id must be a positive integer")
	}

	return carID, nil
}
