package repairparts

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

	parts, err := h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to fetch repair parts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(parts); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateRepairPartInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.DeliveryDate = strings.TrimSpace(input.DeliveryDate)

	if input.Name == "" || input.Quantity < 0 {
		http.Error(w, "name is required and quantity must be 0 or greater", http.StatusBadRequest)
		return
	}

	if input.DeliveryDate != "" {
		if _, err := time.Parse("2006-01-02", input.DeliveryDate); err != nil {
			http.Error(w, "delivery_date must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	part, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create repair part", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(part); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
