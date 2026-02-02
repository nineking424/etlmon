package path

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// MockPathsRepository is a mock implementation of the PathsRepository for testing
type MockPathsRepository struct {
	savedStats []*models.PathStats
	saveError  error
}

func (m *MockPathsRepository) SavePathStats(ctx context.Context, stats *models.PathStats) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.savedStats = append(m.savedStats, stats)
	return nil
}

func (m *MockPathsRepository) GetLatestPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return m.savedStats, nil
}

func (m *MockPathsRepository) GetPathStats(ctx context.Context, path string) (*models.PathStats, error) {
	for _, stats := range m.savedStats {
		if stats.Path == path {
			return stats, nil
		}
	}
	return nil, fmt.Errorf("path not found")
}

// setupTestDir creates a temporary directory structure for testing
func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "etlmon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test structure:
	// tmpDir/
	//   dir1/
	//     file1.txt
	//     file2.log
	//     subdir1/
	//       file3.txt
	//   dir2/
	//     file4.tmp
	//   file5.txt

	dirs := []string{
		filepath.Join(tmpDir, "dir1"),
		filepath.Join(tmpDir, "dir1", "subdir1"),
		filepath.Join(tmpDir, "dir2"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	files := []string{
		filepath.Join(tmpDir, "dir1", "file1.txt"),
		filepath.Join(tmpDir, "dir1", "file2.log"),
		filepath.Join(tmpDir, "dir1", "subdir1", "file3.txt"),
		filepath.Join(tmpDir, "dir2", "file4.tmp"),
		filepath.Join(tmpDir, "file5.txt"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	return tmpDir
}

func TestPathScanner_ScanOnce_CountsFilesCorrectly(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)
	if err != nil {
		t.Fatalf("ScanPath() error = %v", err)
	}

	// Should count 5 files total
	expectedFiles := int64(5)
	if stats.FileCount != expectedFiles {
		t.Errorf("FileCount = %d, want %d", stats.FileCount, expectedFiles)
	}

	// Verify status is OK
	if stats.Status != "OK" {
		t.Errorf("Status = %s, want OK", stats.Status)
	}

	// Verify path is set
	if stats.Path != tmpDir {
		t.Errorf("Path = %s, want %s", stats.Path, tmpDir)
	}
}

func TestPathScanner_ScanOnce_CountsDirsCorrectly(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)
	if err != nil {
		t.Fatalf("ScanPath() error = %v", err)
	}

	// Should count 3 directories: dir1, dir1/subdir1, dir2
	expectedDirs := int64(3)
	if stats.DirCount != expectedDirs {
		t.Errorf("DirCount = %d, want %d", stats.DirCount, expectedDirs)
	}
}

func TestPathScanner_ScanOnce_RespectsMaxDepth(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     1, // Only scan immediate children
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)
	if err != nil {
		t.Fatalf("ScanPath() error = %v", err)
	}

	// With max_depth=1, should only count:
	// - file5.txt (1 file)
	// - dir1, dir2 (2 directories)
	// Should NOT count files inside dir1 or dir1/subdir1
	expectedFiles := int64(1)
	if stats.FileCount != expectedFiles {
		t.Errorf("FileCount with max_depth=1: got %d, want %d", stats.FileCount, expectedFiles)
	}

	expectedDirs := int64(2)
	if stats.DirCount != expectedDirs {
		t.Errorf("DirCount with max_depth=1: got %d, want %d", stats.DirCount, expectedDirs)
	}
}

func TestPathScanner_ScanOnce_AppliesExcludePatterns(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Exclude:      []string{"*.tmp", "*.log"},
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)
	if err != nil {
		t.Fatalf("ScanPath() error = %v", err)
	}

	// Should count 3 files (excluding file2.log and file4.tmp)
	expectedFiles := int64(3)
	if stats.FileCount != expectedFiles {
		t.Errorf("FileCount with excludes = %d, want %d", stats.FileCount, expectedFiles)
	}
}

func TestPathScanner_ScanOnce_RecordsDuration(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)
	if err != nil {
		t.Fatalf("ScanPath() error = %v", err)
	}

	// Should have a non-negative scan duration (can be 0 for very fast scans)
	if stats.ScanDurationMs < 0 {
		t.Errorf("ScanDurationMs = %d, want >= 0", stats.ScanDurationMs)
	}

	// Verify timestamp is recent
	if time.Since(stats.CollectedAt) > 1*time.Second {
		t.Error("CollectedAt timestamp is not recent")
	}
}

func TestPathScanner_ScanOnce_TimesOut_SetsErrorStatus(t *testing.T) {
	// Create a large directory that will take time to scan
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      1 * time.Nanosecond, // Extremely short timeout to force timeout
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	stats, err := scanner.ScanPath(ctx, cfg)

	// Should return stats with ERROR status (not a Go error)
	if err != nil {
		t.Fatalf("ScanPath() should not return error, got: %v", err)
	}

	if stats.Status != "ERROR" {
		t.Errorf("Status = %s, want ERROR", stats.Status)
	}

	if stats.ErrorMessage == "" {
		t.Error("ErrorMessage should not be empty when timeout occurs")
	}
}

func TestPathScanner_ScanOnce_SkipsIfAlreadyScanning(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	// Mark as scanning
	scanner.mu.Lock()
	scanner.scanning[tmpDir] = true
	scanner.mu.Unlock()

	// Try to scan again
	ctx := context.Background()
	_, err := scanner.ScanPath(ctx, cfg)

	// Should skip and return error
	if err == nil {
		t.Error("ScanPath() should return error when already scanning")
	}

	// Cleanup
	scanner.mu.Lock()
	scanner.scanning[tmpDir] = false
	scanner.mu.Unlock()
}

func TestPathScanner_TriggerScan_ScansSpecifiedPaths(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 1 * time.Minute,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx := context.Background()
	err := scanner.TriggerScan(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("TriggerScan() error = %v", err)
	}

	// Verify scan was performed and saved
	if len(repo.savedStats) != 1 {
		t.Errorf("Expected 1 saved stats, got %d", len(repo.savedStats))
	}

	if repo.savedStats[0].Path != tmpDir {
		t.Errorf("Saved path = %s, want %s", repo.savedStats[0].Path, tmpDir)
	}
}

func TestPathScanner_Start_ScansAtInterval(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := &MockPathsRepository{}
	cfg := PathConfig{
		Path:         tmpDir,
		ScanInterval: 100 * time.Millisecond,
		MaxDepth:     10,
		Timeout:      30 * time.Second,
	}

	scanner := NewPathScanner(repo, []PathConfig{cfg})

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	scanner.Start(ctx)

	// Wait for context to complete
	<-ctx.Done()
	scanner.Stop()

	// Should have scanned at least 2 times (initial + 1-2 intervals in 350ms)
	if len(repo.savedStats) < 2 {
		t.Errorf("Expected at least 2 scans, got %d", len(repo.savedStats))
	}
}
