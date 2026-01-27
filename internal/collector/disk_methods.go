package collector

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// CommandTimeout is the default timeout for external commands
const CommandTimeout = 30 * time.Second

// DiskUsage holds disk usage information for a path
type DiskUsage struct {
	Path        string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
	Device      string
	FSType      string
}

// DiskMethodCollector is the interface for disk collection methods
type DiskMethodCollector interface {
	Collect(ctx context.Context, path string) (*DiskUsage, error)
	Name() string
}

// StatsCollector uses gopsutil (syscall.Statfs) to collect disk metrics
type StatsCollector struct{}

// NewStatsCollector creates a new stats collector
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{}
}

// Name returns the collector name
func (c *StatsCollector) Name() string {
	return "stats"
}

// Collect gathers disk usage using gopsutil
func (c *StatsCollector) Collect(ctx context.Context, path string) (*DiskUsage, error) {
	usage, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("stats collection failed for %s: %w", path, err)
	}

	// Get partition info for device and fstype
	partitions, _ := disk.PartitionsWithContext(ctx, false)
	var device, fstype string
	for _, p := range partitions {
		if p.Mountpoint == path {
			device = p.Device
			fstype = p.Fstype
			break
		}
	}

	return &DiskUsage{
		Path:        path,
		Total:       usage.Total,
		Used:        usage.Used,
		Free:        usage.Free,
		UsedPercent: usage.UsedPercent,
		Device:      device,
		FSType:      fstype,
	}, nil
}

// DFCollector uses the df command to collect disk metrics
type DFCollector struct{}

// NewDFCollector creates a new df collector
func NewDFCollector() *DFCollector {
	return &DFCollector{}
}

// Name returns the collector name
func (c *DFCollector) Name() string {
	return "df"
}

// Collect gathers disk usage using the df command
func (c *DFCollector) Collect(ctx context.Context, path string) (*DiskUsage, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, CommandTimeout)
	defer cancel()

	var cmd *exec.Cmd
	var blockMultiplier uint64 = 1

	if runtime.GOOS == "darwin" {
		// macOS: df -b returns 512-byte blocks
		cmd = exec.CommandContext(ctx, "df", "-b", path)
		blockMultiplier = 512
	} else {
		// Linux: df -B1 returns bytes directly
		cmd = exec.CommandContext(ctx, "df", "-B1", path)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("df command failed for %s: %w", path, err)
	}

	return parseDFOutput(string(output), path, blockMultiplier)
}

// parseDFOutput parses df command output
func parseDFOutput(output, path string, blockMultiplier uint64) (*DiskUsage, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty df output")
	}

	// Parse data line
	if !scanner.Scan() {
		return nil, fmt.Errorf("no data in df output")
	}

	line := scanner.Text()
	fields := strings.Fields(line)

	// df output format: Filesystem Blocks Used Available Capacity Mounted
	// Minimum 6 fields expected
	if len(fields) < 6 {
		return nil, fmt.Errorf("unexpected df output format: %s", line)
	}

	device := fields[0]

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing total blocks: %w", err)
	}
	total *= blockMultiplier

	used, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing used blocks: %w", err)
	}
	used *= blockMultiplier

	free, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing free blocks: %w", err)
	}
	free *= blockMultiplier

	var usedPercent float64
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100
	}

	return &DiskUsage{
		Path:        path,
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
		Device:      device,
		FSType:      "", // df doesn't provide fstype
	}, nil
}

// DUCollector uses the du command to collect disk metrics
type DUCollector struct{}

// NewDUCollector creates a new du collector
func NewDUCollector() *DUCollector {
	return &DUCollector{}
}

// Name returns the collector name
func (c *DUCollector) Name() string {
	return "du"
}

// Collect gathers disk usage using the du command (recursive)
func (c *DUCollector) Collect(ctx context.Context, path string) (*DiskUsage, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, CommandTimeout)
	defer cancel()

	// du -sb: summarize, bytes (works on both Linux and macOS with coreutils)
	// On macOS without coreutils, fall back to du -s and multiply by 512
	var cmd *exec.Cmd
	var blockMultiplier uint64 = 1

	if runtime.GOOS == "darwin" {
		// Try du -s first (512-byte blocks on macOS)
		cmd = exec.CommandContext(ctx, "du", "-s", path)
		blockMultiplier = 512
	} else {
		// Linux: du -sb for bytes
		cmd = exec.CommandContext(ctx, "du", "-sb", path)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("du command failed for %s: %w", path, err)
	}

	return parseDUOutput(string(output), path, blockMultiplier)
}

// parseDUOutput parses du command output
func parseDUOutput(output, path string, blockMultiplier uint64) (*DiskUsage, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty du output")
	}

	// du output format: SIZE\tPATH
	fields := strings.Fields(output)
	if len(fields) < 1 {
		return nil, fmt.Errorf("unexpected du output format: %s", output)
	}

	used, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing du size: %w", err)
	}
	used *= blockMultiplier

	// du only gives us used space, not total
	// For du method, we can't calculate percentage without knowing total
	return &DiskUsage{
		Path:        path,
		Total:       0, // Unknown for du
		Used:        used,
		Free:        0, // Unknown for du
		UsedPercent: 0, // Cannot calculate without total
		Device:      "",
		FSType:      "",
	}, nil
}

// GetMethodCollector returns the appropriate collector for a method
func GetMethodCollector(method string) DiskMethodCollector {
	switch method {
	case "df":
		return NewDFCollector()
	case "du":
		return NewDUCollector()
	default:
		return NewStatsCollector()
	}
}
