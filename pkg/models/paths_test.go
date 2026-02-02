package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPathStats_JSONMarshaling(t *testing.T) {
	now := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	stats := PathStats{
		Path:           "/data/logs",
		FileCount:      1500,
		DirCount:       50,
		ScanDurationMs: 250,
		Status:         "OK",
		ErrorMessage:   "",
		CollectedAt:    now,
	}

	// Marshal to JSON
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded PathStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if decoded.Path != stats.Path {
		t.Errorf("Path: got %s, want %s", decoded.Path, stats.Path)
	}
	if decoded.FileCount != stats.FileCount {
		t.Errorf("FileCount: got %d, want %d", decoded.FileCount, stats.FileCount)
	}
	if decoded.DirCount != stats.DirCount {
		t.Errorf("DirCount: got %d, want %d", decoded.DirCount, stats.DirCount)
	}
	if decoded.ScanDurationMs != stats.ScanDurationMs {
		t.Errorf("ScanDurationMs: got %d, want %d", decoded.ScanDurationMs, stats.ScanDurationMs)
	}
	if decoded.Status != stats.Status {
		t.Errorf("Status: got %s, want %s", decoded.Status, stats.Status)
	}
	if decoded.ErrorMessage != stats.ErrorMessage {
		t.Errorf("ErrorMessage: got %s, want %s", decoded.ErrorMessage, stats.ErrorMessage)
	}
	if !decoded.CollectedAt.Equal(stats.CollectedAt) {
		t.Errorf("CollectedAt: got %v, want %v", decoded.CollectedAt, stats.CollectedAt)
	}
}

func TestPathStats_JSONMarshaling_WithError(t *testing.T) {
	now := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	stats := PathStats{
		Path:           "/data/restricted",
		FileCount:      0,
		DirCount:       0,
		ScanDurationMs: 10,
		Status:         "ERROR",
		ErrorMessage:   "permission denied",
		CollectedAt:    now,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded PathStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ErrorMessage != stats.ErrorMessage {
		t.Errorf("ErrorMessage: got %s, want %s", decoded.ErrorMessage, stats.ErrorMessage)
	}
}

func TestPathStats_JSONTags(t *testing.T) {
	now := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	stats := PathStats{
		Path:           "/data/input",
		FileCount:      100,
		DirCount:       10,
		ScanDurationMs: 150,
		Status:         "SCANNING",
		CollectedAt:    now,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify JSON structure matches expected field names
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	expectedFields := []string{
		"path",
		"file_count",
		"dir_count",
		"scan_duration_ms",
		"status",
		"collected_at",
	}

	for _, field := range expectedFields {
		if _, exists := raw[field]; !exists {
			t.Errorf("Expected JSON field %s not found", field)
		}
	}

	// error_message should be omitted when empty
	if stats.ErrorMessage == "" {
		if _, exists := raw["error_message"]; exists {
			t.Errorf("error_message should be omitted when empty")
		}
	}
}

func TestPathStats_StatusValues(t *testing.T) {
	validStatuses := []string{"OK", "SCANNING", "ERROR"}

	for _, status := range validStatuses {
		stats := PathStats{
			Path:        "/test",
			Status:      status,
			CollectedAt: time.Now(),
		}

		data, err := json.Marshal(stats)
		if err != nil {
			t.Fatalf("Failed to marshal status %s: %v", status, err)
		}

		var decoded PathStats
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal status %s: %v", status, err)
		}

		if decoded.Status != status {
			t.Errorf("Status: got %s, want %s", decoded.Status, status)
		}
	}
}
