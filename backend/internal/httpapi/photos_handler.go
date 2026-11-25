package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Ssnakerss/E-Wardrobe/internal/photos"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// photoHandler holds dependencies for photo HTTP handlers.
type photoHandler struct {
	svc      *photos.Service
	pipeline *photos.UploadPipeline
	pool     *database.Pool
}

// newPhotoHandler creates a new photo HTTP handler.
func newPhotoHandler(svc *photos.Service, pipeline *photos.UploadPipeline, pool *database.Pool) *photoHandler {
	return &photoHandler{svc: svc, pipeline: pipeline, pool: pool}
}

// registerPhotoRoutes registers photo routes on the given router.
func (h *photoHandler) registerPhotoRoutes(r chi.Router) {
	r.Post("/items/{itemId}/photos", h.createPhoto)
	r.Put("/items/{itemId}/photos/{photoId}/content", h.uploadPhotoContent)
	r.Get("/photos/{photoId}/content", h.getPhotoContent)
	r.Delete("/photos/{photoId}", h.deletePhoto)
	r.Put("/items/{itemId}/photos/order", h.updatePhotoOrder)
}

// createPhoto handles POST /items/{itemId}/photos.
func (h *photoHandler) createPhoto(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	// Read file with size limit
	r.Body = http.MaxBytesReader(w, r.Body, photos.MaxUploadSize+1024)
	if err := r.ParseMultipartForm(photos.MaxUploadSize + 1024); err != nil {
		RespondError(w, BadRequest("file too large"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			RespondError(w, BadRequest("missing file"))
			return
		}
		RespondError(w, BadRequest("failed to read file"))
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		RespondError(w, photos.ValidationError("only JPEG and PNG images are supported"))
		return
	}

	// Get sort order from form or use last + 1
	sortOrder := int32(0)
	if sortOrderStr := r.FormValue("sort_order"); sortOrderStr != "" {
		if v, err := strconv.ParseInt(sortOrderStr, 10, 32); err == nil {
			sortOrder = int32(v)
		}
	}

	var photo photos.Photo
	photo, err = h.pipeline.UploadPhoto(r.Context(), userID, itemID, file, sortOrder)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondCreated(w, photo)
}

// uploadPhotoContent handles PUT /items/{itemId}/photos/{photoId}/content.
func (h *photoHandler) uploadPhotoContent(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	photoID, err := uuid.Parse(chi.URLParam(r, "photoId"))
	if err != nil {
		RespondError(w, BadRequest("invalid photo_id"))
		return
	}

	// Read file with size limit
	r.Body = http.MaxBytesReader(w, r.Body, photos.MaxUploadSize+1024)
	file, err := io.ReadAll(io.LimitReader(r.Body, photos.MaxUploadSize+1))
	if err != nil {
		RespondError(w, BadRequest("file too large"))
		return
	}

	fileReader := bytes.NewReader(file)

	var photo photos.Photo
	photo, err = h.pipeline.UploadContent(r.Context(), userID, photoID, fileReader)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, photo)
}

// getPhotoContent handles GET /photos/{photoId}/content.
func (h *photoHandler) getPhotoContent(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	photoID, err := uuid.Parse(chi.URLParam(r, "photoId"))
	if err != nil {
		RespondError(w, BadRequest("invalid photo_id"))
		return
	}

	variant := r.URL.Query().Get("variant")
	if variant != "thumbnail" && variant != "original" && variant != "" {
		RespondError(w, BadRequest("invalid variant"))
		return
	}

	rc, err := h.svc.GetContent(r.Context(), nil, photoID, userID, variant)
	if err != nil {
		RespondError(w, NotFound("photo not found"))
		return
	}
	defer rc.Close()

	// Get photo metadata for content type
	photo, err := h.svc.GetByID(r.Context(), nil, photoID, userID)
	if err != nil {
		RespondError(w, NotFound("photo not found"))
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if photo.MimeType != nil {
		w.Header().Set("Content-Type", *photo.MimeType)
	}

	if photo.SizeBytes != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*photo.SizeBytes, 10))
	}

	// Serve content directly without seeking
	http.ServeContent(w, r, "", photo.UpdatedAt, nopSeeker{rc})
}

// nopSeeker wraps an io.ReadCloser to implement io.ReadSeeker
// by always returning SeekEnd with offset 0.
type nopSeeker struct {
	io.ReadCloser
}

func (s nopSeeker) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// deletePhoto handles DELETE /photos/{photoId}.
func (h *photoHandler) deletePhoto(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	photoID, err := uuid.Parse(chi.URLParam(r, "photoId"))
	if err != nil {
		RespondError(w, BadRequest("invalid photo_id"))
		return
	}

	// Get current version
	photo, err := h.svc.GetByID(r.Context(), nil, photoID, userID)
	if err != nil {
		RespondError(w, NotFound("photo not found"))
		return
	}

	err = h.svc.SoftDelete(r.Context(), nil, photoID, userID, photo.Version)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondNoContent(w)
}

// updatePhotoOrder handles PUT /items/{itemId}/photos/order.
func (h *photoHandler) updatePhotoOrder(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	var orders []photos.PhotoOrder
	if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
		RespondError(w, BadRequest("invalid JSON"))
		return
	}

	if len(orders) == 0 {
		RespondError(w, BadRequest("empty orders"))
		return
	}

	err = h.svc.UpdatePhotoOrderBatch(r.Context(), nil, userID, itemID, orders)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondNoContent(w)
}
