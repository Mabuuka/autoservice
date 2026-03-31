package employees

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

	employees, err := h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "failed to fetch employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(employees); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateEmployeeInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	input.Specialty = strings.TrimSpace(input.Specialty)
	input.Phone = strings.TrimSpace(input.Phone)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.PersonnelNumber <= 0 || input.Specialty == "" || input.Phone == "" || input.FullName == "" {
		http.Error(w, "personnel_number, specialty, phone and full_name are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	employee, err := h.Repo.Create(ctx, input)
	if err != nil {
		http.Error(w, "failed to create employee", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(employee); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
