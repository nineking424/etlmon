package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFilesystemUsage_JSONMarshaling(t *testing.T) {
	now := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	fs := FilesystemUsage{
		MountPoint:  "/data",
		TotalBytes:  1000000000,
		UsedBytes:   600000000,
		AvailBytes:  400000000,
		UsedPercent: 60.0,
		CollectedAt: now,
	}

	// Marshal to JSON
	data, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded FilesystemUsage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if decoded.MountPoint != fs.MountPoint {
		t.Errorf("MountPoint: got %s, want %s", decoded.MountPoint, fs.MountPoint)
	}
	if decoded.TotalBytes != fs.TotalBytes {
		t.Errorf("TotalBytes: got %d, want %d", decoded.TotalBytes, fs.TotalBytes)
	}
	if decoded.UsedBytes != fs.UsedBytes {
		t.Errorf("UsedBytes: got %d, want %d", decoded.UsedBytes, fs.UsedBytes)
	}
	if decoded.AvailBytes != fs.AvailBytes {
		t.Errorf("AvailBytes: got %d, want %d", decoded.AvailBytes, fs.AvailBytes)
	}
	if decoded.UsedPercent != fs.UsedPercent {
		t.Errorf("UsedPercent: got %f, want %f", decoded.UsedPercent, fs.UsedPercent)
	}
	if !decoded.CollectedAt.Equal(fs.CollectedAt) {
		t.Errorf("CollectedAt: got %v, want %v", decoded.CollectedAt, fs.CollectedAt)
	}
}

func TestFilesystemUsage_JSONTags(t *testing.T) {
	now := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	fs := FilesystemUsage{
		MountPoint:  "/",
		TotalBytes:  500000000000,
		UsedBytes:   300000000000,
		AvailBytes:  200000000000,
		UsedPercent: 60.0,
		CollectedAt: now,
	}

	data, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify JSON structure matches expected field names
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	expectedFields := []string{
		"mount_point",
		"total_bytes",
		"used_bytes",
		"avail_bytes",
		"used_percent",
		"collected_at",
	}

	for _, field := range expectedFields {
		if _, exists := raw[field]; !exists {
			t.Errorf("Expected JSON field %s not found", field)
		}
	}
}
