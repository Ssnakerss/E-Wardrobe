package httpapi

import (
	"encoding/json"
	"net/http"
)

// RespondJSON writes a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// RespondOK writes a 200 OK JSON response.
func RespondOK(w http.ResponseWriter, v any) {
	RespondJSON(w, http.StatusOK, v)
}

// RespondCreated writes a 201 Created JSON response.
func RespondCreated(w http.ResponseWriter, v any) {
	RespondJSON(w, http.StatusCreated, v)
}

// RespondNoContent writes a 204 No Content response.
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondError writes a standardized error response.
func RespondError(w http.ResponseWriter, err error) {
	WriteError(w, err)
}
