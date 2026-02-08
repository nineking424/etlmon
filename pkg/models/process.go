package models

import "time"

// ProcessInfo represents a monitored process's statistics
type ProcessInfo struct {
	PID         int       `json:"pid"`
	Name        string    `json:"name"`
	User        string    `json:"user"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemRSS      int64     `json:"mem_rss"`       // bytes
	Status      string    `json:"status"`         // running, sleeping, zombie, stopped
	Elapsed     string    `json:"elapsed"`        // human-readable elapsed time
	CollectedAt time.Time `json:"collected_at"`
}
