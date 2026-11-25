package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Ssnakerss/E-Wardrobe/internal/items"
	"github.com/Ssnakerss/E-Wardrobe/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// itemHandler holds dependencies for item HTTP handlers.
type itemHandler struct {
	svc  *items.Service
	pool *database.Pool
}

// newItemHandler creates a new items HTTP handler.
func newItemHandler(svc *items.Service, pool *database.Pool) *itemHandler {
	return &itemHandler{svc: svc, pool: pool}
}

// registerRoutes registers item routes on the given router.
func (h *itemHandler) registerRoutes(r chi.Router) {
	r.Get("/storage-locations", h.getStorageLocations)
	r.Post("/items", h.createItem)
	r.Get("/items", h.listItems)
	r.Get("/items/{itemId}", h.getItem)
	r.Patch("/items/{itemId}", h.updateItem)
	r.Delete("/items/{itemId}", h.deleteItem)
	r.Post("/items/{itemId}/archive", h.toggleArchive)
}

// createItem handles POST /items.
func (h *itemHandler) createItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)

	var raw rawItemInput
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		RespondError(w, BadRequest("invalid JSON"))
		return
	}

	input, parseErr := raw.ToItemInput()
	if parseErr != nil {
		RespondError(w, parseErr)
		return
	}

	var item items.Item
	var err error
	item, err = h.svc.Create(r.Context(), h.pool, userID, input)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondCreated(w, item)
}

// rawItemInput is the raw JSON input for creating an item.
type rawItemInput struct {
	CategoryID       string  `json:"category_id"`
	Name             string  `json:"name"`
	Brand            *string `json:"brand"`
	Description      *string `json:"description"`
	PurchaseYear     *string `json:"purchase_year"`
	PurchasePrice    *string `json:"purchase_price"`
	PriceCurrency    *string `json:"price_currency"`
	PurchasePlace    *string `json:"purchase_place"`
	ConditionRating  *string `json:"condition_rating"`
	PreferenceRating *string `json:"preference_rating"`
	LastWornOn       *string `json:"last_worn_on"`
	StorageLocation  *string `json:"storage_location"`
	Season           *string `json:"season"`
	Color            *string `json:"color"`
	Size             *string `json:"size"`
	Status           string  `json:"status"`
}

// ToItemInput converts raw input to domain items.ItemInput.
func (r *rawItemInput) ToItemInput() (items.ItemInput, *AppError) {
	catID, err := uuid.Parse(r.CategoryID)
	if err != nil {
		return items.ItemInput{}, ValidationError("invalid category_id")
	}

	input := items.ItemInput{
		CategoryID:      catID,
		Name:            r.Name,
		Brand:           r.Brand,
		Description:     r.Description,
		StorageLocation: r.StorageLocation,
		Season:          r.Season,
		Color:           r.Color,
		Size:            r.Size,
		Status:          r.Status,
	}

	if r.PurchaseYear != nil {
		v, err := strconv.ParseInt(*r.PurchaseYear, 10, 16)
		if err != nil {
			return items.ItemInput{}, ValidationError("invalid purchase_year")
		}
		v16 := int16(v)
		input.PurchaseYear = &v16
	}

	if r.PurchasePrice != nil {
		v, err := strconv.ParseInt(*r.PurchasePrice, 10, 64)
		if err != nil {
			return items.ItemInput{}, ValidationError("invalid purchase_price")
		}
		input.PurchasePrice = &v
	}

	if r.PriceCurrency != nil {
		input.PriceCurrency = r.PriceCurrency
	}

	if r.PurchasePlace != nil {
		input.PurchasePlace = r.PurchasePlace
	}

	if r.ConditionRating != nil {
		v, err := strconv.ParseInt(*r.ConditionRating, 10, 16)
		if err != nil {
			return items.ItemInput{}, ValidationError("invalid condition_rating")
		}
		v16 := int16(v)
		input.ConditionRating = &v16
	}

	if r.PreferenceRating != nil {
		v, err := strconv.ParseInt(*r.PreferenceRating, 10, 16)
		if err != nil {
			return items.ItemInput{}, ValidationError("invalid preference_rating")
		}
		v16 := int16(v)
		input.PreferenceRating = &v16
	}

	if r.LastWornOn != nil && *r.LastWornOn != "" {
		t, err := time.Parse("2006-01-02", *r.LastWornOn)
		if err != nil {
			return items.ItemInput{}, ValidationError("invalid last_worn_on")
		}
		input.LastWornOn = &t
	}

	return input, nil
}

// listItems handles GET /items.
func (h *itemHandler) listItems(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)

	opts := items.ListOptions{
		UserID: userID,
		Limit:  50,
	}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			opts.CategoryID = &id
		}
	}
	if query := r.URL.Query().Get("q"); query != "" {
		opts.Query = &query
	}
	if status := r.URL.Query().Get("status"); status != "" {
		opts.Status = &status
	}
	if season := r.URL.Query().Get("season"); season != "" {
		opts.Season = &season
	}
	if archived := r.URL.Query().Get("archived"); archived == "true" {
		opts.Archived = true
	}

	itemsList, err := h.svc.List(r.Context(), h.pool, opts)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]any{
		"items":       itemsList,
		"next_cursor": (*string)(nil),
	})
}

// getItem handles GET /items/{itemId}.
func (h *itemHandler) getItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	item, err := h.svc.GetByID(r.Context(), h.pool, itemID, userID)
	if err != nil {
		RespondError(w, NotFound("item not found"))
		return
	}

	RespondJSON(w, http.StatusOK, item)
}

// updateItem handles PATCH /items/{itemId}.
func (h *itemHandler) updateItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	var raw rawPatchInput
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		RespondError(w, BadRequest("invalid JSON"))
		return
	}

	patch, err := raw.ToItemPatch(itemID)
	if err != nil {
		RespondError(w, err)
		return
	}

	item, err := h.svc.Update(r.Context(), h.pool, userID, patch)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, item)
}

// rawPatchInput is the raw JSON input for updating an item.
type rawPatchInput struct {
	BaseVersion      *int64  `json:"base_version"`
	CategoryID       *string `json:"category_id"`
	Name             *string `json:"name"`
	Brand            *string `json:"brand"`
	Description      *string `json:"description"`
	PurchaseYear     *string `json:"purchase_year"`
	PurchasePrice    *string `json:"purchase_price"`
	PriceCurrency    *string `json:"price_currency"`
	PurchasePlace    *string `json:"purchase_place"`
	ConditionRating  *string `json:"condition_rating"`
	PreferenceRating *string `json:"preference_rating"`
	LastWornOn       *string `json:"last_worn_on"`
	StorageLocation  *string `json:"storage_location"`
	Season           *string `json:"season"`
	Color            *string `json:"color"`
	Size             *string `json:"size"`
	Status           *string `json:"status"`
}

// ToItemPatch converts raw input to domain items.ItemPatch.
func (r *rawPatchInput) ToItemPatch(itemID uuid.UUID) (items.ItemPatch, *AppError) {
	patch := items.ItemPatch{
		ID:          itemID,
		BaseVersion: 0,
	}

	if r.BaseVersion != nil {
		patch.BaseVersion = *r.BaseVersion
	}

	if r.CategoryID != nil {
		if id, err := uuid.Parse(*r.CategoryID); err == nil {
			patch.CategoryID = &id
		} else {
			return items.ItemPatch{}, ValidationError("invalid category_id")
		}
	}

	patch.Name = r.Name
	patch.Brand = r.Brand
	patch.Description = r.Description
	patch.PurchasePlace = r.PurchasePlace
	patch.StorageLocation = r.StorageLocation
	patch.Color = r.Color
	patch.Size = r.Size
	patch.Status = r.Status

	if r.PurchaseYear != nil {
		v, err := strconv.ParseInt(*r.PurchaseYear, 10, 16)
		if err != nil {
			return items.ItemPatch{}, ValidationError("invalid purchase_year")
		}
		v16 := int16(v)
		patch.PurchaseYear = &v16
	}

	if r.PurchasePrice != nil {
		v, err := strconv.ParseInt(*r.PurchasePrice, 10, 64)
		if err != nil {
			return items.ItemPatch{}, ValidationError("invalid purchase_price")
		}
		patch.PurchasePrice = &v
	}

	if r.PriceCurrency != nil {
		patch.PriceCurrency = r.PriceCurrency
	}

	if r.ConditionRating != nil {
		v, err := strconv.ParseInt(*r.ConditionRating, 10, 16)
		if err != nil {
			return items.ItemPatch{}, ValidationError("invalid condition_rating")
		}
		v16 := int16(v)
		patch.ConditionRating = &v16
	}

	if r.PreferenceRating != nil {
		v, err := strconv.ParseInt(*r.PreferenceRating, 10, 16)
		if err != nil {
			return items.ItemPatch{}, ValidationError("invalid preference_rating")
		}
		v16 := int16(v)
		patch.PreferenceRating = &v16
	}

	if r.LastWornOn != nil && *r.LastWornOn != "" {
		t, err := time.Parse("2006-01-02", *r.LastWornOn)
		if err != nil {
			return items.ItemPatch{}, ValidationError("invalid last_worn_on")
		}
		patch.LastWornOn = &t
	}

	if r.Season != nil {
		patch.Season = r.Season
	}

	return patch, nil
}

// deleteItem handles DELETE /items/{itemId}.
func (h *itemHandler) deleteItem(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	item, err := h.svc.GetByID(r.Context(), h.pool, itemID, userID)
	if err != nil {
		RespondError(w, NotFound("item not found"))
		return
	}

	err = h.svc.SoftDelete(r.Context(), h.pool, itemID, userID, item.Version)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondNoContent(w)
}

// toggleArchive handles POST /items/{itemId}/archive.
func (h *itemHandler) toggleArchive(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		RespondError(w, BadRequest("invalid item_id"))
		return
	}

	item, err := h.svc.GetByID(r.Context(), h.pool, itemID, userID)
	if err != nil {
		RespondError(w, NotFound("item not found"))
		return
	}

	if item.ArchivedAt != nil {
		item, err = h.svc.Unarchive(r.Context(), h.pool, itemID, userID, item.Version)
	} else {
		item, err = h.svc.Archive(r.Context(), h.pool, itemID, userID, item.Version)
	}

	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, item)
}

// getStorageLocations handles GET /storage-locations.
func (h *itemHandler) getStorageLocations(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromContext(r)

	query := r.URL.Query().Get("q")
	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(v)
		}
	}

	suggestions, err := h.svc.ListStorageLocationSuggestions(r.Context(), h.pool, userID, query, limit)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]any{
		"suggestions": suggestions,
	})
}
