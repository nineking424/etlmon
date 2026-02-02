package models

import "time"

// FilesystemUsage represents disk usage statistics for a mount point
type FilesystemUsage struct {
	MountPoint  string    `json:"mount_point"`  // Mount point path (e.g., "/data")
	TotalBytes  uint64    `json:"total_bytes"`  // Total filesystem size in bytes
	UsedBytes   uint64    `json:"used_bytes"`   // Used space in bytes
	AvailBytes  uint64    `json:"avail_bytes"`  // Available space in bytes
	UsedPercent float64   `json:"used_percent"` // Usage percentage (0-100)
	CollectedAt time.Time `json:"collected_at"` // When this metric was collected
}
