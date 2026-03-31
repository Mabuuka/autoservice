package cars

import (
	"context"
	"encoding/json"
	"net/http"
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

	cars, err := h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to fetch cars", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(cars); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateCarInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.Brand = strings.TrimSpace(input.Brand)
	input.PlateNumber = strings.TrimSpace(input.PlateNumber)
	input.Color = strings.TrimSpace(input.Color)

	if input.OwnerID <= 0 || input.Brand == "" || input.PlateNumber == "" || input.ManufactureYear <= 0 || input.Color == "" {
		http.Error(w, "owner_id, brand, plate_number, manufacture_year and color are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	car, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create car", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(car); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
