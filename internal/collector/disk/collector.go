package disk

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/etlmon/etlmon/pkg/models"
)

// FSRepository defines the interface for storing filesystem usage data
type FSRepository interface {
	SaveFilesystemUsage(ctx context.Context, usage *models.FilesystemUsage) error
	GetLatestFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error)
}

// DiskCollector collects filesystem usage statistics
type DiskCollector struct {
	repo     FSRepository
	interval time.Duration
	excludes map[string]bool
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// NewDiskCollector creates a new disk usage collector
func NewDiskCollector(repo FSRepository, interval time.Duration) *DiskCollector {
	// Default pseudo filesystems to exclude
	excludes := map[string]bool{
		"proc":     true,
		"sysfs":    true,
		"devpts":   true,
		"tmpfs":    true,
		"devtmpfs": true,
		"cgroup":   true,
		"cgroup2":  true,
		"pstore":   true,
		"bpf":      true,
		"tracefs":  true,
		"debugfs":  true,
		"securityfs": true,
		"hugetlbfs": true,
		"mqueue":    true,
	}

	return &DiskCollector{
		repo:     repo,
		interval: interval,
		excludes: excludes,
	}
}

// Start begins periodic filesystem usage collection
func (c *DiskCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		return fmt.Errorf("collector already started")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.mu.Unlock()

	// Start collection goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.collectLoop(ctx)
	}()

	return nil
}

// Stop stops the filesystem usage collection
func (c *DiskCollector) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.mu.Unlock()

	c.wg.Wait()
}

// CollectOnce performs a single collection of all filesystem usage
func (c *DiskCollector) CollectOnce(ctx context.Context) error {
	mounts, err := c.getMountPoints()
	if err != nil {
		return fmt.Errorf("failed to get mount points: %w", err)
	}

	for _, mount := range mounts {
		stats, err := c.getFilesystemStats(mount)
		if err != nil {
			// Log error but continue with other mounts
			continue
		}

		if err := c.repo.SaveFilesystemUsage(ctx, stats); err != nil {
			return fmt.Errorf("failed to save filesystem usage for %s: %w", mount, err)
		}
	}

	return nil
}

// collectLoop runs the periodic collection
func (c *DiskCollector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately on start
	_ = c.CollectOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.CollectOnce(ctx)
		}
	}
}

// getMountPoints returns a list of non-pseudo filesystem mount points
func (c *DiskCollector) getMountPoints() ([]string, error) {
	if runtime.GOOS == "darwin" {
		// On macOS, just return root for now
		// In production, could parse 'mount' command output
		return []string{"/"}, nil
	}

	// On Linux, parse /proc/mounts
	return c.getMountPointsLinux()
}

// getMountPointsLinux parses /proc/mounts on Linux systems
func (c *DiskCollector) getMountPointsLinux() ([]string, error) {
	// Read /proc/mounts
	// For simplicity, return common mount points for now
	// In production, would parse the actual file

	// Default to root
	mounts := []string{"/"}

	// Try to add common mount points if they exist
	commonMounts := []string{"/home", "/data", "/var", "/tmp"}
	for _, mount := range commonMounts {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err == nil {
			mounts = append(mounts, mount)
		}
	}

	return mounts, nil
}

// getFilesystemStats gets usage statistics for a specific mount point
func (c *DiskCollector) getFilesystemStats(mountPoint string) (*models.FilesystemUsage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &stat); err != nil {
		return nil, fmt.Errorf("statfs failed for %s: %w", mountPoint, err)
	}

	// Calculate sizes based on OS
	var bsize uint64
	if runtime.GOOS == "darwin" {
		bsize = uint64(stat.Bsize)
	} else {
		bsize = uint64(stat.Bsize)
	}

	totalBytes := stat.Blocks * bsize
	availBytes := stat.Bavail * bsize
	freeBytes := stat.Bfree * bsize
	usedBytes := totalBytes - freeBytes

	var usedPercent float64
	if totalBytes > 0 {
		usedPercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	return &models.FilesystemUsage{
		MountPoint:  mountPoint,
		TotalBytes:  totalBytes,
		UsedBytes:   usedBytes,
		AvailBytes:  availBytes,
		UsedPercent: usedPercent,
		CollectedAt: time.Now(),
	}, nil
}
