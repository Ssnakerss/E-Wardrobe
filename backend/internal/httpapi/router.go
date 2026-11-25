package httpapi

import (
	"net/http"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(pool *database.Pool) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health endpoints (no auth)
	r.Get("/health/live", healthLive)
	r.Get("/health/ready", healthReady)

	// Auth endpoints (no auth)
	r.Post("/api/v1/auth/invitations/validate", authValidateInvitation)
	r.Post("/api/v1/auth/register", authRegister)
	r.Post("/api/v1/auth/login", authLogin)
	r.Post("/api/v1/auth/logout", authLogout)
	r.Get("/api/v1/auth/csrf", authCSRF)
	r.Get("/api/v1/auth/me", authMe)

	// Recovery endpoints (no auth)
	r.Post("/api/v1/auth/recovery/validate", recoveryValidate)
	r.Post("/api/v1/auth/recovery/complete", recoveryComplete)

	// Protected endpoints (require auth)
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return authMiddleware(pool, next)
		})

		// Categories
		r.Get("/api/v1/categories", listCategories)
		r.Post("/api/v1/categories", createCategory)
		r.Patch("/api/v1/categories/{categoryId}", updateCategory)
		r.Delete("/api/v1/categories/{categoryId}", deleteCategory)

		// Items
		r.Get("/api/v1/items", listItems)
		r.Post("/api/v1/items", createItem)
		r.Get("/api/v1/items/{itemId}", getItem)
		r.Patch("/api/v1/items/{itemId}", updateItem)
		r.Delete("/api/v1/items/{itemId}", deleteItem)
		r.Post("/api/v1/items/{itemId}/archive", archiveItem)
		r.Get("/api/v1/storage-locations", storageLocations)

		// Photos
		r.Post("/api/v1/items/{itemId}/photos", createPhoto)
		r.Put("/api/v1/items/{itemId}/photos/{photoId}/content", uploadPhotoContent)
		r.Get("/api/v1/photos/{photoId}/content", getPhotoContent)
		r.Delete("/api/v1/photos/{photoId}", deletePhoto)
		r.Put("/api/v1/items/{itemId}/photos/order", updatePhotoOrder)

		// Sync
		r.Post("/api/v1/sync/push", syncPush)
		r.Get("/api/v1/sync/pull", syncPull)
	})

	return r
}

// Placeholder handlers — to be replaced with actual implementations.

func healthLive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func healthReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func authValidateInvitation(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func authRegister(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusCreated, map[string]any{"message": "placeholder"})
}

func authLogin(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func authLogout(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func authCSRF(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"csrf_token": "placeholder"})
}

func authMe(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func recoveryValidate(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func recoveryComplete(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func listCategories(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"categories": []any{}})
}

func createCategory(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusCreated, map[string]any{"message": "placeholder"})
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func listItems(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"items": []any{}, "next_cursor": (*string)(nil)})
}

func createItem(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusCreated, map[string]any{"message": "placeholder"})
}

func getItem(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func archiveItem(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func storageLocations(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"suggestions": []string{}})
}

func createPhoto(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusCreated, map[string]any{"message": "placeholder"})
}

func uploadPhotoContent(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func getPhotoContent(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func deletePhoto(w http.ResponseWriter, r *http.Request) {
	RespondNoContent(w)
}

func updatePhotoOrder(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"message": "placeholder"})
}

func syncPush(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{"results": []any{}})
}

func syncPull(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]any{
		"changes":   []any{},
		"next_since": int64(0),
		"has_more":  false,
	})
}
