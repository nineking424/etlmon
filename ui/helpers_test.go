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

func TestFormatGauge_FormatsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		width    int
		expected string
	}{
		{
			name:     "zero percent",
			percent:  0,
			width:    10,
			expected: "[░░░░░░░░░░]   0.0%",
		},
		{
			name:     "50 percent",
			percent:  50,
			width:    10,
			expected: "[█████░░░░░]  50.0%",
		},
		{
			name:     "100 percent",
			percent:  100,
			width:    10,
			expected: "[██████████] 100.0%",
		},
		{
			name:     "89.1 percent",
			percent:  89.1,
			width:    20,
			expected: "[█████████████████░░░]  89.1%",
		},
		{
			name:     "negative clamped to zero",
			percent:  -5,
			width:    5,
			expected: "[░░░░░]   0.0%",
		},
		{
			name:     "over 100 clamped",
			percent:  150,
			width:    5,
			expected: "[█████] 100.0%",
		},
		{
			name:     "minimum width",
			percent:  50,
			width:    1,
			expected: "[█░░]  50.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatGauge(tt.percent, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatNumber_FormatsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		n        int64
		expected string
	}{
		{
			name:     "zero",
			n:        0,
			expected: "0",
		},
		{
			name:     "small number",
			n:        999,
			expected: "999",
		},
		{
			name:     "thousands",
			n:        1234,
			expected: "1,234",
		},
		{
			name:     "millions",
			n:        1234567,
			expected: "1,234,567",
		},
		{
			name:     "negative",
			n:        -1234,
			expected: "-1,234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNumber(tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}
