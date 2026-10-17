package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inventory/internal/repository"
	"inventory/internal/service"
)

func TestWriteErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"validation", &service.ValidationError{Fields: map[string]string{"price": "bad"}}, http.StatusUnprocessableEntity},
		{"duplicate sku", &service.ValidationError{Fields: map[string]string{"sku": "dup"}, Conflict: true}, http.StatusConflict},
		{"not found", repository.ErrNotFound, http.StatusNotFound},
		{"insufficient stock", repository.ErrInsufficientStock, http.StatusUnprocessableEntity},
		{"unexpected", http.ErrAbortHandler, http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeError(rec, tc.err)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
				t.Fatalf("expected a JSON error body, got %q", rec.Body.String())
			}
		})
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"x","category":"Stationery"}`))
	var in struct {
		Name string `json:"name"`
	}
	if err := decode(httptest.NewRecorder(), req, &in); err == nil {
		t.Fatal("expected an error for an unknown field")
	}
}

func TestPathIDRejectsBadValues(t *testing.T) {
	for _, v := range []string{"abc", "0", "-3", ""} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", v)
		if _, ok := pathID(req); ok {
			t.Errorf("pathID accepted %q", v)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", "7")
	if id, ok := pathID(req); !ok || id != 7 {
		t.Errorf("pathID(7) = %d, %v", id, ok)
	}
}
