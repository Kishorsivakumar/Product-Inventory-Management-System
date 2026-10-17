package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"inventory/internal/repository"
	"inventory/internal/service"
)

type errorBody struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// writeError maps domain errors to HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		status := http.StatusUnprocessableEntity // 422
		if ve.Conflict {
			status = http.StatusConflict // 409
		}
		writeJSON(w, status, errorBody{Error: "Validation failed", Fields: ve.Fields})
	case errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody{Error: "Resource not found"})
	case errors.Is(err, repository.ErrInsufficientStock):
		writeJSON(w, http.StatusUnprocessableEntity, errorBody{
			Error:  "Insufficient stock",
			Fields: map[string]string{"quantity": "Cannot remove more than the available stock"},
		})
	default:
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "Internal server error"})
	}
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, errorBody{Error: msg})
}
