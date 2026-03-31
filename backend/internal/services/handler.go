package services

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

	services, err := h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to fetch services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateServiceInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" || input.PriceRub < 0 || input.RegularDiscountPercent < 0 || input.RegularDiscountPercent > 100 {
		http.Error(w, "name is required, price_rub must be >= 0, regular_discount_percent must be between 0 and 100", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	service, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(service); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
