package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBytes_FormatsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1.00 KB",
		},
		{
			name:     "megabytes",
			bytes:    1024 * 1024,
			expected: "1.00 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1024 * 1024 * 1024,
			expected: "1.00 GB",
		},
		{
			name:     "terabytes",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.00 TB",
		},
		{
			name:     "partial kilobytes",
			bytes:    1536,
			expected: "1.50 KB",
		},
		{
			name:     "partial gigabytes",
			bytes:    2560 * 1024 * 1024,
			expected: "2.50 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes_HandlesZero(t *testing.T) {
	result := FormatBytes(0)
	assert.Equal(t, "0 B", result)
}

func TestFormatDuration_FormatsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		expected string
	}{
		{
			name:     "milliseconds",
			ms:       500,
			expected: "500ms",
		},
		{
			name:     "seconds",
			ms:       1500,
			expected: "1.5s",
		},
		{
			name:     "minutes",
			ms:       90000,
			expected: "1m30s",
		},
		{
			name:     "zero",
			ms:       0,
			expected: "0ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.ms)
			assert.Equal(t, tt.expected, result)
		})
	}
}
