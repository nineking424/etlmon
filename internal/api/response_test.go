package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/etlmon/etlmon/pkg/models"
)

func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	writeJSON(w, http.StatusOK, data)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestWriteJSON_SetsStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	writeJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestWriteJSON_EncodesBody(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	writeJSON(w, http.StatusOK, data)

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("expected message 'test', got '%s'", result["message"])
	}
}

func TestWriteError_ReturnsErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("something went wrong")

	writeError(w, http.StatusInternalServerError, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var result models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if result.Error != "something went wrong" {
		t.Errorf("expected error message 'something went wrong', got '%s'", result.Error)
	}
}

func TestWriteError_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("test error")

	writeError(w, http.StatusBadRequest, err)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
