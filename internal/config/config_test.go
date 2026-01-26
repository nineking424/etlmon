package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create temp config file
	content := `
interval: 10s
resources:
  - cpu
  - memory
  - disk
windows:
  - 1m
  - 5m
  - 1h
aggregations:
  - avg
  - max
  - min
  - last
database:
  path: ./etlmon.db
`
	tmpFile := createTempConfig(t, content)
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	
	if cfg.Interval != 10*time.Second {
		t.Errorf("Interval = %v, want %v", cfg.Interval, 10*time.Second)
	}
	if len(cfg.Resources) != 3 {
		t.Errorf("len(Resources) = %d, want 3", len(cfg.Resources))
	}
	if len(cfg.Windows) != 3 {
		t.Errorf("len(Windows) = %d, want 3", len(cfg.Windows))
	}
	if len(cfg.Aggregations) != 4 {
		t.Errorf("len(Aggregations) = %d, want 4", len(cfg.Aggregations))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `invalid: yaml: content: [[[`
	tmpFile := createTempConfig(t, content)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
}

func TestValidate_MissingInterval(t *testing.T) {
	cfg := &Config{
		Resources:    []string{"cpu"},
		Windows:      []string{"1m"},
		Aggregations: []string{"avg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for missing interval")
	}
}

func TestValidate_MissingResources(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{},
		Windows:      []string{"1m"},
		Aggregations: []string{"avg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for missing resources")
	}
}

func TestValidate_InvalidResource(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"invalid_resource"},
		Windows:      []string{"1m"},
		Aggregations: []string{"avg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid resource")
	}
}

func TestValidate_MissingWindows(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"cpu"},
		Windows:      []string{},
		Aggregations: []string{"avg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for missing windows")
	}
}

func TestValidate_InvalidWindow(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"cpu"},
		Windows:      []string{"invalid"},
		Aggregations: []string{"avg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid window")
	}
}

func TestValidate_MissingAggregations(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"cpu"},
		Windows:      []string{"1m"},
		Aggregations: []string{},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for missing aggregations")
	}
}

func TestValidate_InvalidAggregation(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"cpu"},
		Windows:      []string{"1m"},
		Aggregations: []string{"invalid_agg"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() expected error for invalid aggregation")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		Resources:    []string{"cpu", "memory", "disk"},
		Windows:      []string{"1m", "5m", "1h"},
		Aggregations: []string{"avg", "max", "min", "last"},
		Database: DatabaseConfig{
			Path: "./test.db",
		},
	}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestParseWindow_ValidFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1m", time.Minute},
		{"5m", 5 * time.Minute},
		{"1h", time.Hour},
		{"30s", 30 * time.Second},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseWindow(tt.input)
			if err != nil {
				t.Errorf("ParseWindow(%q) error = %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseWindow(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseWindow_InvalidFormats(t *testing.T) {
	invalids := []string{"", "abc", "1x", "-1m"}
	
	for _, input := range invalids {
		t.Run(input, func(t *testing.T) {
			_, err := ParseWindow(input)
			if err == nil {
				t.Errorf("ParseWindow(%q) expected error", input)
			}
		})
	}
}

func TestGetWindowDurations(t *testing.T) {
	cfg := &Config{
		Windows: []string{"1m", "5m", "1h"},
	}
	
	durations, err := cfg.GetWindowDurations()
	if err != nil {
		t.Fatalf("GetWindowDurations() error = %v", err)
	}
	
	expected := []time.Duration{time.Minute, 5 * time.Minute, time.Hour}
	if len(durations) != len(expected) {
		t.Fatalf("len(durations) = %d, want %d", len(durations), len(expected))
	}
	
	for i, d := range durations {
		if d != expected[i] {
			t.Errorf("durations[%d] = %v, want %v", i, d, expected[i])
		}
	}
}

// Helper function
func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return tmpFile
}
