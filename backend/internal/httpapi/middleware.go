package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/Ssnakerss/E-Wardrobe/internal/auth"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/google/uuid"
)

// ContextKey defines types for context values.
type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
)

// authMiddleware extracts the user ID from the request and stores it in the context.
// It supports both Bearer token (mobile) and cookie-based (web) authentication.
func authMiddleware(pool *database.Pool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			token = extractCookieToken(r)
		}

		if token == "" {
			WriteError(w, Unauthorized("missing authentication"))
			return
		}

		// Look up session and extract user ID
		userID, err := auth.GetUserIDFromSessionPool(r.Context(), pool, token)
		if err != nil {
			WriteError(w, Unauthorized("invalid session"))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

// extractCookieToken extracts the session cookie.
func extractCookieToken(r *http.Request) string {
	cookie, err := r.Cookie("wardrobe_session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// requireAuth wraps a handler and ensures the user is authenticated.
// The user ID is available via UserIDKey in the request context.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
		if !ok || userID == uuid.Nil {
			WriteError(w, Unauthorized("missing authentication"))
			return
		}
		next(w, r)
	}
}

// getUserIDFromContext extracts the user ID from the request context.
func getUserIDFromContext(r *http.Request) uuid.UUID {
	userID, _ := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID
}

// GetUserIDFromContext is the exported version of getUserIDFromContext.
func GetUserIDFromContext(r *http.Request) uuid.UUID {
	return getUserIDFromContext(r)
}
