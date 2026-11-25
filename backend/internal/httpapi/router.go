package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/categories"
	"github.com/Ssnakerss/E-Wardrobe/internal/changejournal"
	"github.com/Ssnakerss/E-Wardrobe/internal/items"
	"github.com/Ssnakerss/E-Wardrobe/internal/photos"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/storage"
	"github.com/Ssnakerss/E-Wardrobe/internal/sync"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	Pool      *database.Pool
	Store     storage.PhotoStorage
	TmpDir    string
	Journal   *changejournal.Writer
	CatSvc    *categories.Service
	CatRepo   *categories.Repository
	ItemSvc   *items.Service
	PhotoSvc  *photos.Service
	PhotoRepo *photos.Repository
}

// NewRouter creates and configures the HTTP router.
func NewRouter(cfg RouterConfig) http.Handler {
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

	// Create sync service
	syncRepo := sync.NewRepository(cfg.Pool)
	syncSvc := sync.NewService(cfg.Pool, syncRepo, cfg.Journal, cfg.CatSvc, cfg.ItemSvc, cfg.PhotoSvc)

	// Protected endpoints (require auth)
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return authMiddleware(cfg.Pool, next)
		})

		// Categories (placeholder)
		r.Get("/api/v1/categories", listCategories)
		r.Post("/api/v1/categories", createCategory)
		r.Patch("/api/v1/categories/{categoryId}", updateCategory)
		r.Delete("/api/v1/categories/{categoryId}", deleteCategory)

		// Items
		itemHandler := newItemHandler(cfg.ItemSvc, cfg.Pool)
		itemHandler.registerRoutes(r)

		// Photos
		photoPipeline := photos.NewUploadPipeline(cfg.PhotoRepo, cfg.Store, cfg.Journal, cfg.Pool)
		photoHandler := newPhotoHandler(cfg.PhotoSvc, photoPipeline, cfg.Pool)
		photoHandler.registerPhotoRoutes(r)

		// Sync
		r.Post("/api/v1/sync/push", syncPushHandler(syncSvc))
		r.Get("/api/v1/sync/pull", syncPullHandler(syncSvc))
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

// syncPushHandler returns an HTTP handler for sync push.
func syncPushHandler(svc *sync.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromContext(r)

		var req sync.PushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, BadRequest("invalid request body"))
			return
		}

		result, err := svc.Push(r.Context(), userID, req)
		if err != nil {
			WriteError(w, err)
			return
		}

		RespondJSON(w, http.StatusOK, result)
	}
}

// syncPullHandler returns an HTTP handler for sync pull.
func syncPullHandler(svc *sync.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromContext(r)

		sinceStr := r.URL.Query().Get("since")
		limitStr := r.URL.Query().Get("limit")

		since := int64(0)
		if sinceStr != "" {
			val, err := strconv.ParseInt(sinceStr, 10, 64)
			if err != nil {
				WriteError(w, BadRequest("invalid since parameter"))
				return
			}
			since = val
		}

		limit := int32(200)
		if limitStr != "" {
			val, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				WriteError(w, BadRequest("invalid limit parameter"))
				return
			}
			limit = int32(val)
		}

		result, err := svc.Pull(r.Context(), userID, since, limit)
		if err != nil {
			WriteError(w, err)
			return
		}

		RespondJSON(w, http.StatusOK, result)
	}
}
