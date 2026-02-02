package models

import "time"

// PathStats represents file/directory count statistics for a monitored path
type PathStats struct {
	Path           string    `json:"path"`                      // Path being monitored (e.g., "/data/logs")
	FileCount      int64     `json:"file_count"`                // Number of files found
	DirCount       int64     `json:"dir_count"`                 // Number of directories found
	ScanDurationMs int64     `json:"scan_duration_ms"`          // How long the scan took in milliseconds
	Status         string    `json:"status"`                    // Current status: OK, SCANNING, ERROR
	ErrorMessage   string    `json:"error_message,omitempty"`   // Error details if status is ERROR
	CollectedAt    time.Time `json:"collected_at"`              // When this scan completed
}
