package models

import "time"

// LogEntry represents a single log line from a monitored file
type LogEntry struct {
	ID        int64     `json:"id"`
	LogName   string    `json:"log_name"`
	LogPath   string    `json:"log_path"`
	Line      string    `json:"line"`
	CreatedAt time.Time `json:"created_at"`
}

// LogConfig defines a log file to monitor
type LogConfig struct {
	Name     string `yaml:"name" json:"name"`
	Path     string `yaml:"path" json:"path"`
	MaxLines int    `yaml:"max_lines" json:"max_lines"`
}
