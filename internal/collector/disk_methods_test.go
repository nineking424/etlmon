package collector

import (
	"context"
	"testing"
	"time"
)

func TestStatsCollector_Name(t *testing.T) {
	c := NewStatsCollector()
	if c.Name() != "stats" {
		t.Errorf("Name() = %s, want stats", c.Name())
	}
}

func TestStatsCollector_Collect(t *testing.T) {
	c := NewStatsCollector()
	ctx := context.Background()

	// Test with root path (should always exist)
	usage, err := c.Collect(ctx, "/")
	if err != nil {
		t.Fatalf("Collect() failed: %v", err)
	}

	if usage.Path != "/" {
		t.Errorf("Path = %s, want /", usage.Path)
	}
	if usage.Total == 0 {
		t.Error("Total = 0, want > 0")
	}
	if usage.UsedPercent < 0 || usage.UsedPercent > 100 {
		t.Errorf("UsedPercent = %f, want 0-100", usage.UsedPercent)
	}
}

func TestStatsCollector_CollectInvalidPath(t *testing.T) {
	c := NewStatsCollector()
	ctx := context.Background()

	_, err := c.Collect(ctx, "/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestDFCollector_Name(t *testing.T) {
	c := NewDFCollector()
	if c.Name() != "df" {
		t.Errorf("Name() = %s, want df", c.Name())
	}
}

func TestDFCollector_Collect(t *testing.T) {
	c := NewDFCollector()
	ctx := context.Background()

	usage, err := c.Collect(ctx, "/")
	if err != nil {
		t.Fatalf("Collect() failed: %v", err)
	}

	if usage.Path != "/" {
		t.Errorf("Path = %s, want /", usage.Path)
	}
	if usage.Total == 0 {
		t.Error("Total = 0, want > 0")
	}
}

func TestDUCollector_Name(t *testing.T) {
	c := NewDUCollector()
	if c.Name() != "du" {
		t.Errorf("Name() = %s, want du", c.Name())
	}
}

func TestDUCollector_Collect(t *testing.T) {
	c := NewDUCollector()
	ctx := context.Background()

	// Use /tmp which should exist and be readable
	usage, err := c.Collect(ctx, "/tmp")
	if err != nil {
		t.Fatalf("Collect() failed: %v", err)
	}

	if usage.Path != "/tmp" {
		t.Errorf("Path = %s, want /tmp", usage.Path)
	}
	// du only returns used bytes, total should be 0
	if usage.Total != 0 {
		t.Errorf("Total = %d, want 0 for du method", usage.Total)
	}
}

func TestParseDFOutput_Linux(t *testing.T) {
	// Simulated df -B1 output from Linux
	output := `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1    499963174912 392020549632 107942625280  79% /`

	usage, err := parseDFOutput(output, "/", 1)
	if err != nil {
		t.Fatalf("parseDFOutput() failed: %v", err)
	}

	if usage.Device != "/dev/sda1" {
		t.Errorf("Device = %s, want /dev/sda1", usage.Device)
	}
	if usage.Total != 499963174912 {
		t.Errorf("Total = %d, want 499963174912", usage.Total)
	}
	if usage.Used != 392020549632 {
		t.Errorf("Used = %d, want 392020549632", usage.Used)
	}
	if usage.Free != 107942625280 {
		t.Errorf("Free = %d, want 107942625280", usage.Free)
	}
}

func TestParseDFOutput_MacOS(t *testing.T) {
	// Simulated df -b output from macOS (512-byte blocks)
	output := `Filesystem    512-blocks      Used Available Capacity  Mounted on
/dev/disk1s1   976490576 765664160 210826416    79%    /`

	usage, err := parseDFOutput(output, "/", 512)
	if err != nil {
		t.Fatalf("parseDFOutput() failed: %v", err)
	}

	if usage.Device != "/dev/disk1s1" {
		t.Errorf("Device = %s, want /dev/disk1s1", usage.Device)
	}
	// 976490576 * 512 = 499963174912
	if usage.Total != 976490576*512 {
		t.Errorf("Total = %d, want %d", usage.Total, 976490576*512)
	}
}

func TestParseDUOutput(t *testing.T) {
	output := "1234567\t/some/path"

	usage, err := parseDUOutput(output, "/some/path", 1)
	if err != nil {
		t.Fatalf("parseDUOutput() failed: %v", err)
	}

	if usage.Used != 1234567 {
		t.Errorf("Used = %d, want 1234567", usage.Used)
	}
	if usage.Total != 0 {
		t.Errorf("Total = %d, want 0 for du", usage.Total)
	}
}

func TestParseDUOutput_MacOS(t *testing.T) {
	// macOS du -s output (512-byte blocks)
	output := "2469134\t/some/path"

	usage, err := parseDUOutput(output, "/some/path", 512)
	if err != nil {
		t.Fatalf("parseDUOutput() failed: %v", err)
	}

	// 2469134 * 512 = 1264196608
	if usage.Used != 2469134*512 {
		t.Errorf("Used = %d, want %d", usage.Used, 2469134*512)
	}
}

func TestGetMethodCollector(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{"stats", "stats"},
		{"df", "df"},
		{"du", "du"},
		{"unknown", "stats"}, // Default to stats
		{"", "stats"},        // Default to stats
	}

	for _, tt := range tests {
		collector := GetMethodCollector(tt.method)
		if collector.Name() != tt.want {
			t.Errorf("GetMethodCollector(%s).Name() = %s, want %s", tt.method, collector.Name(), tt.want)
		}
	}
}

func TestCommandTimeout(t *testing.T) {
	if CommandTimeout != 30*time.Second {
		t.Errorf("CommandTimeout = %v, want 30s", CommandTimeout)
	}
}
