package users

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"autoservice/backend/internal/api"
	"autoservice/backend/internal/auth"
)

type Handler struct {
	Repo     *Repository
	Sessions *auth.SessionManager
}

func NewHandler(repo *Repository, sessions *auth.SessionManager) *Handler {
	return &Handler{Repo: repo, Sessions: sessions}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	normalizeRegisterInput(&input)
	if err := validateRegisterInput(input); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		api.Internal(w, "Failed to prepare user password.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.Repo.CreateClient(ctx, input, passwordHash)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "User with the same email, full_name or phone already exists.")
		default:
			api.Internal(w, "Failed to register user.")
		}
		return
	}

	authToken := h.Sessions.BuildToken(user.UserID)
	h.Sessions.SetAuthCookie(w, user.UserID)

	profile, err := h.Repo.GetProfileByUserID(ctx, user.UserID)
	if err != nil {
		api.Internal(w, "User was created, but profile loading failed.")
		return
	}

	_ = api.WriteJSON(w, http.StatusCreated, map[string]any{
		"message":    "User registered successfully.",
		"data":       profile,
		"auth_token": authToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	normalizeLoginInput(&input)
	if err := validateLoginInput(input); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.Repo.GetAuthUserByEmail(ctx, input.Email)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			api.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
		default:
			api.Internal(w, "Failed to login user.")
		}
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, input.Password); err != nil {
		api.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
		return
	}

	if err := h.Repo.UpdateLastLogin(ctx, user.UserID); err != nil {
		api.Internal(w, "Failed to update login info.")
		return
	}

	authToken := h.Sessions.BuildToken(user.UserID)
	h.Sessions.SetAuthCookie(w, user.UserID)

	profile, err := h.Repo.GetProfileByUserID(ctx, user.UserID)
	if err != nil {
		api.Internal(w, "User authenticated, but profile loading failed.")
		return
	}

	_ = api.WriteJSON(w, http.StatusOK, map[string]any{
		"message":    "User logged in successfully.",
		"data":       profile,
		"auth_token": authToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Sessions.ClearAuthCookie(w)
	_ = api.WriteSuccessMessage(w, http.StatusOK, "User logged out successfully.")
}

func (h *Handler) GetCurrentProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	profile, err := h.Repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Current user was not found.")
		default:
			api.Internal(w, "Failed to fetch current profile.")
		}
		return
	}

	_ = api.WriteData(w, http.StatusOK, profile)
}

func (h *Handler) UpdateCurrentProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	var input UpdateProfileInput
	if err := api.DecodeJSONBody(r, &input); err != nil {
		api.WriteRequestError(w, err)
		return
	}

	normalizeUpdateProfileInput(&input)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	currentUser, err := h.Repo.GetAuthUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Current user was not found.")
		default:
			api.Internal(w, "Failed to load current user.")
		}
		return
	}

	if err := validateUpdateProfileInput(currentUser.Role, input); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	profile, err := h.Repo.UpdateProfile(ctx, userID, input)
	if err != nil {
		switch {
		case api.IsUniqueViolation(err):
			api.Conflict(w, "Profile with the same email, full_name or phone already exists.")
		case api.IsCheckViolation(err):
			api.BadRequest(w, "preferred_entrypoint contains an invalid value.")
		default:
			api.Internal(w, "Failed to update profile.")
		}
		return
	}

	_ = api.WriteUpdated(w, "Profile updated successfully.", profile)
}

func normalizeRegisterInput(input *RegisterInput) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Password = strings.TrimSpace(input.Password)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.DriverLicense = strings.TrimSpace(input.DriverLicense)
}

func normalizeLoginInput(input *LoginInput) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Password = strings.TrimSpace(input.Password)
}

func normalizeUpdateProfileInput(input *UpdateProfileInput) {
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.PreferredEntrypoint = strings.TrimSpace(input.PreferredEntrypoint)
	input.DriverLicense = strings.TrimSpace(input.DriverLicense)
}

func validateRegisterInput(input RegisterInput) error {
	if input.Email == "" || input.Password == "" || input.FullName == "" || input.Phone == "" || input.DriverLicense == "" {
		return errors.New("Fields email, password, full_name, phone and driver_license are required.")
	}

	if !strings.Contains(input.Email, "@") {
		return errors.New("email must be valid.")
	}

	if len(input.Password) < 8 {
		return errors.New("password must contain at least 8 characters.")
	}

	return nil
}

func validateLoginInput(input LoginInput) error {
	if input.Email == "" || input.Password == "" {
		return errors.New("Fields email and password are required.")
	}

	return nil
}

func validateUpdateProfileInput(role string, input UpdateProfileInput) error {
	if input.FullName == "" || input.Phone == "" || input.PreferredEntrypoint == "" {
		return errors.New("Fields full_name, phone and preferred_entrypoint are required.")
	}

	switch input.PreferredEntrypoint {
	case "profile", "cars", "orders":
	default:
		return errors.New("preferred_entrypoint must be one of: profile, cars, orders.")
	}

	if role == "client" && input.DriverLicense == "" {
		return errors.New("Field driver_license is required for client profile update.")
	}

	return nil
}
