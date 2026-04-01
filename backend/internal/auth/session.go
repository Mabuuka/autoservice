package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"autoservice/backend/internal/api"
	"autoservice/backend/internal/config"
)

var ErrUnauthorized = errors.New("unauthorized")

type SessionManager struct {
	CookieName string
	Secret     string
	Secure     bool
	TTL        time.Duration
}

func NewSessionManager(cfg config.Config) *SessionManager {
	return &SessionManager{
		CookieName: cfg.AuthCookieName,
		Secret:     cfg.AuthSecret,
		Secure:     cfg.AuthCookieSecure,
		TTL:        time.Duration(cfg.AuthCookieTTL) * time.Hour,
	}
}

func (s *SessionManager) SetAuthCookie(w http.ResponseWriter, userID int64) {
	expiresAt := time.Now().Add(s.TTL)
	value := s.buildRawToken(userID, expiresAt)

	http.SetCookie(w, &http.Cookie{
		Name:     s.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.Secure,
		Expires:  expiresAt,
		MaxAge:   int(s.TTL.Seconds()),
	})
}

func (s *SessionManager) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.Secure,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func (s *SessionManager) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := s.ReadUserID(r)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	}
}

func (s *SessionManager) ReadUserID(r *http.Request) (int64, error) {
	if token := bearerTokenFromRequest(r); token != "" {
		return s.ReadUserIDFromToken(token)
	}

	if token := strings.TrimSpace(r.Header.Get("X-Auth-Token")); token != "" {
		return s.ReadUserIDFromToken(token)
	}

	cookie, err := r.Cookie(s.CookieName)
	if err != nil {
		return 0, ErrUnauthorized
	}

	return s.ReadUserIDFromToken(cookie.Value)
}

func (s *SessionManager) ReadUserIDFromToken(token string) (int64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, ErrUnauthorized
	}

	if userID, err := s.parseRawToken(token); err == nil {
		return userID, nil
	}

	rawToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, ErrUnauthorized
	}

	return s.parseRawToken(string(rawToken))
}

func (s *SessionManager) BuildToken(userID int64) string {
	expiresAt := time.Now().Add(s.TTL)
	return s.buildRawToken(userID, expiresAt)
}

func (s *SessionManager) buildRawToken(userID int64, expiresAt time.Time) string {
	payload := fmt.Sprintf("%d|%d", userID, expiresAt.Unix())
	signature := s.sign(payload)
	return payload + "|" + signature
}

func (s *SessionManager) parseRawToken(token string) (int64, error) {
	parts := strings.Split(token, "|")
	if len(parts) != 3 {
		return 0, ErrUnauthorized
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrUnauthorized
	}

	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}

	expectedSignature := s.sign(parts[0] + "|" + parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return 0, ErrUnauthorized
	}

	if time.Now().Unix() > expiresUnix {
		return 0, ErrUnauthorized
	}

	return userID, nil
}

func (s *SessionManager) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}


func bearerTokenFromRequest(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return ""
	}

	const prefix = "Bearer "
	if len(authorization) < len(prefix) || !strings.EqualFold(authorization[:len(prefix)], prefix) {
		return ""
	}

	return strings.TrimSpace(authorization[len(prefix):])
}
