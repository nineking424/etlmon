package collector

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Test Collector Interface
func TestCPUCollector_Type(t *testing.T) {
	c := NewCPUCollector()
	if c.Type() != "cpu" {
		t.Errorf("Type() = %s, want cpu", c.Type())
	}
}

func TestMemoryCollector_Type(t *testing.T) {
	c := NewMemoryCollector()
	if c.Type() != "memory" {
		t.Errorf("Type() = %s, want memory", c.Type())
	}
}

func TestDiskCollector_Type(t *testing.T) {
	c := NewDiskCollector()
	if c.Type() != "disk" {
		t.Errorf("Type() = %s, want disk", c.Type())
	}
}

// Test CPU Collector
func TestCPUCollector_Collect(t *testing.T) {
	c := NewCPUCollector()
	ctx := context.Background()

	metrics, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Collect() returned no metrics")
	}

	// Check that we have usage_percent metric
	found := false
	for _, m := range metrics {
		if m.Name == "usage_percent" {
			found = true
			if m.Value < 0 || m.Value > 100 {
				t.Errorf("CPU usage_percent = %v, want 0-100", m.Value)
			}
		}
	}
	if !found {
		t.Error("Collect() missing usage_percent metric")
	}
}

// Test Memory Collector
func TestMemoryCollector_Collect(t *testing.T) {
	c := NewMemoryCollector()
	ctx := context.Background()

	metrics, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Collect() returned no metrics")
	}

	// Check that we have usage_percent metric
	found := false
	for _, m := range metrics {
		if m.Name == "usage_percent" {
			found = true
			if m.Value < 0 || m.Value > 100 {
				t.Errorf("Memory usage_percent = %v, want 0-100", m.Value)
			}
		}
	}
	if !found {
		t.Error("Collect() missing usage_percent metric")
	}
}

// Test Disk Collector
func TestDiskCollector_Collect(t *testing.T) {
	c := NewDiskCollector()
	ctx := context.Background()

	metrics, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	// Note: Disk collector may return 0 metrics if all disks are pseudo-FS
	// Just verify no error and valid values if any metrics
	for _, m := range metrics {
		if m.Value < 0 || m.Value > 100 && m.Name == "usage_percent" {
			t.Errorf("Disk %s = %v, want 0-100", m.Name, m.Value)
		}
	}
}

// Test Manager
func TestManager_Register(t *testing.T) {
	m := NewManager(1 * time.Second)
	c := NewCPUCollector()

	m.Register(c)

	// Verify collector was registered
	if len(m.collectors) != 1 {
		t.Errorf("len(collectors) = %d, want 1", len(m.collectors))
	}
}

func TestManager_RegisterMultiple(t *testing.T) {
	m := NewManager(1 * time.Second)

	m.Register(NewCPUCollector())
	m.Register(NewMemoryCollector())
	m.Register(NewDiskCollector())

	if len(m.collectors) != 3 {
		t.Errorf("len(collectors) = %d, want 3", len(m.collectors))
	}
}

func TestManager_Start_ContextCancel(t *testing.T) {
	m := NewManager(100 * time.Millisecond)
	m.Register(NewCPUCollector())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var received []Metric
	var mu sync.Mutex

	handler := func(metrics []Metric) {
		mu.Lock()
		received = append(received, metrics...)
		mu.Unlock()
	}

	done := make(chan struct{})
	go func() {
		m.Start(ctx, handler)
		close(done)
	}()

	// Wait for completion
	select {
	case <-done:
		// Good, Start returned after context cancel
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not return after context cancel")
	}

	mu.Lock()
	count := len(received)
	mu.Unlock()

	// Should have received at least some metrics
	if count == 0 {
		t.Error("No metrics received")
	}
}

func TestManager_CollectOnce(t *testing.T) {
	m := NewManager(1 * time.Second)
	m.Register(NewCPUCollector())
	m.Register(NewMemoryCollector())

	ctx := context.Background()
	metrics, err := m.CollectOnce(ctx)
	if err != nil {
		t.Fatalf("CollectOnce() error = %v", err)
	}

	// Should have metrics from both collectors
	cpuFound := false
	memFound := false
	for _, metric := range metrics {
		if metric.ResourceType == "cpu" {
			cpuFound = true
		}
		if metric.ResourceType == "memory" {
			memFound = true
		}
	}

	if !cpuFound {
		t.Error("No CPU metrics collected")
	}
	if !memFound {
		t.Error("No memory metrics collected")
	}
}

// Test Metric struct
func TestMetric_Fields(t *testing.T) {
	m := Metric{
		ResourceType: "cpu",
		Name:         "usage_percent",
		Value:        45.5,
		Timestamp:    time.Now(),
	}

	if m.ResourceType != "cpu" {
		t.Errorf("ResourceType = %s, want cpu", m.ResourceType)
	}
	if m.Name != "usage_percent" {
		t.Errorf("Name = %s, want usage_percent", m.Name)
	}
	if m.Value != 45.5 {
		t.Errorf("Value = %v, want 45.5", m.Value)
	}
}

// Test pseudo-FS filtering (for disk)
func TestIsPseudoFS(t *testing.T) {
	pseudoTypes := []string{"tmpfs", "devtmpfs", "sysfs", "proc", "overlay", "squashfs"}
	realTypes := []string{"ext4", "xfs", "ntfs", "apfs", "hfs"}

	for _, fs := range pseudoTypes {
		if !isPseudoFS(fs) {
			t.Errorf("isPseudoFS(%s) = false, want true", fs)
		}
	}

	for _, fs := range realTypes {
		if isPseudoFS(fs) {
			t.Errorf("isPseudoFS(%s) = true, want false", fs)
		}
	}
}
