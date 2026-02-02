package disk

import (
	"context"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// MockFSRepository is a mock implementation of the FSRepository for testing
type MockFSRepository struct {
	savedUsage []*models.FilesystemUsage
	saveError  error
}

func (m *MockFSRepository) SaveFilesystemUsage(ctx context.Context, usage *models.FilesystemUsage) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.savedUsage = append(m.savedUsage, usage)
	return nil
}

func (m *MockFSRepository) GetLatestFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return m.savedUsage, nil
}

func TestDiskCollector_getMountPoints_ReturnsNonPseudoMounts(t *testing.T) {
	repo := &MockFSRepository{}
	collector := NewDiskCollector(repo, 1*time.Second)

	mounts, err := collector.getMountPoints()
	if err != nil {
		t.Fatalf("getMountPoints() error = %v", err)
	}

	// Verify we got at least the root mount
	if len(mounts) == 0 {
		t.Fatal("getMountPoints() returned no mounts")
	}

	// Verify pseudo filesystems are excluded
	pseudoFS := map[string]bool{
		"proc":    true,
		"sysfs":   true,
		"devpts":  true,
		"tmpfs":   true,
		"devtmpfs": true,
	}

	for _, mount := range mounts {
		if pseudoFS[mount] {
			t.Errorf("getMountPoints() returned pseudo filesystem: %s", mount)
		}
	}
}

func TestDiskCollector_getFilesystemStats_CalculatesCorrectly(t *testing.T) {
	repo := &MockFSRepository{}
	collector := NewDiskCollector(repo, 1*time.Second)

	// Test with root filesystem (should always exist)
	stats, err := collector.getFilesystemStats("/")
	if err != nil {
		t.Fatalf("getFilesystemStats() error = %v", err)
	}

	// Verify basic calculations
	if stats.TotalBytes == 0 {
		t.Error("TotalBytes should not be zero")
	}

	if stats.UsedBytes > stats.TotalBytes {
		t.Errorf("UsedBytes (%d) should not exceed TotalBytes (%d)", stats.UsedBytes, stats.TotalBytes)
	}

	if stats.AvailBytes > stats.TotalBytes {
		t.Errorf("AvailBytes (%d) should not exceed TotalBytes (%d)", stats.AvailBytes, stats.TotalBytes)
	}

	// Verify percentage calculation
	expectedPercent := float64(stats.UsedBytes) / float64(stats.TotalBytes) * 100
	if stats.UsedPercent < expectedPercent-0.1 || stats.UsedPercent > expectedPercent+0.1 {
		t.Errorf("UsedPercent = %.2f, want approximately %.2f", stats.UsedPercent, expectedPercent)
	}

	// Verify mount point is set
	if stats.MountPoint != "/" {
		t.Errorf("MountPoint = %s, want /", stats.MountPoint)
	}

	// Verify timestamp is recent
	if time.Since(stats.CollectedAt) > 1*time.Second {
		t.Error("CollectedAt timestamp is not recent")
	}
}

func TestDiskCollector_CollectOnce_SavesAllMounts(t *testing.T) {
	repo := &MockFSRepository{}
	collector := NewDiskCollector(repo, 1*time.Second)

	ctx := context.Background()
	err := collector.CollectOnce(ctx)
	if err != nil {
		t.Fatalf("CollectOnce() error = %v", err)
	}

	// Verify at least one mount was saved
	if len(repo.savedUsage) == 0 {
		t.Fatal("CollectOnce() did not save any filesystem usage data")
	}

	// Verify all saved entries have valid data
	for i, usage := range repo.savedUsage {
		if usage.MountPoint == "" {
			t.Errorf("savedUsage[%d].MountPoint is empty", i)
		}
		if usage.TotalBytes == 0 {
			t.Errorf("savedUsage[%d].TotalBytes is zero", i)
		}
		if usage.CollectedAt.IsZero() {
			t.Errorf("savedUsage[%d].CollectedAt is zero", i)
		}
	}
}

func TestDiskCollector_Start_CollectsAtInterval(t *testing.T) {
	repo := &MockFSRepository{}
	interval := 100 * time.Millisecond
	collector := NewDiskCollector(repo, interval)

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	err := collector.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Wait for context to complete
	<-ctx.Done()
	collector.Stop()

	// We should have collected at least 3 times (initial + 2 intervals in 350ms with 100ms interval)
	// Being conservative, check for at least 2 collections
	if len(repo.savedUsage) < 2 {
		t.Errorf("Expected at least 2 collections, got %d", len(repo.savedUsage))
	}
}

func TestDiskCollector_ExcludesPseudoFS(t *testing.T) {
	repo := &MockFSRepository{}
	collector := NewDiskCollector(repo, 1*time.Second)

	ctx := context.Background()
	err := collector.CollectOnce(ctx)
	if err != nil {
		t.Fatalf("CollectOnce() error = %v", err)
	}

	// Verify no pseudo filesystems were saved
	pseudoFS := map[string]bool{
		"proc":    true,
		"sysfs":   true,
		"devpts":  true,
		"tmpfs":   true,
		"devtmpfs": true,
	}

	for _, usage := range repo.savedUsage {
		if pseudoFS[usage.MountPoint] {
			t.Errorf("Pseudo filesystem %s was not excluded", usage.MountPoint)
		}
	}
}

func TestDiskCollector_Stop_StopsCollection(t *testing.T) {
	repo := &MockFSRepository{}
	interval := 50 * time.Millisecond
	collector := NewDiskCollector(repo, interval)

	ctx := context.Background()
	err := collector.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Let it collect a few times
	time.Sleep(150 * time.Millisecond)

	// Stop the collector
	collector.Stop()

	// Record count immediately after stop
	countAfterStop := len(repo.savedUsage)

	// Wait a bit more and verify no new collections
	time.Sleep(200 * time.Millisecond)
	finalCount := len(repo.savedUsage)

	if finalCount != countAfterStop {
		t.Errorf("Collections continued after Stop(). Count at stop: %d, Final: %d", countAfterStop, finalCount)
	}
}
