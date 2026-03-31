package owners

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

	owners, err := h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to fetch owners", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(owners); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateOwnerInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.DriverLicense = strings.TrimSpace(input.DriverLicense)

	if input.FullName == "" || input.Phone == "" || input.DriverLicense == "" {
		http.Error(w, "full_name, phone and driver_license are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	owner, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create owner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(owner); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
