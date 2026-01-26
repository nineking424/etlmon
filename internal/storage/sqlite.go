package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements storage using SQLite
type SQLiteStore struct {
	db     *sql.DB
	path   string
	mu     sync.Mutex
	closed bool
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	// Open database with WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &SQLiteStore{
		db:   db,
		path: dbPath,
	}, nil
}

// Initialize creates the database schema
func (s *SQLiteStore) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}

	return nil
}

// SaveAggregatedMetric saves a single aggregated metric
func (s *SQLiteStore) SaveAggregatedMetric(metric *AggregatedMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
        INSERT INTO aggregated_metrics (timestamp, resource_type, metric_name, aggregated_value, window_size, aggregation_type)
        VALUES (?, ?, ?, ?, ?, ?)
    `, metric.Timestamp, metric.ResourceType, metric.MetricName, metric.AggregatedValue, metric.WindowSize, metric.AggregationType)

	if err != nil {
		return fmt.Errorf("inserting metric: %w", err)
	}

	return nil
}

// SaveBatch saves multiple metrics in a transaction
func (s *SQLiteStore) SaveBatch(metrics []*AggregatedMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT INTO aggregated_metrics (timestamp, resource_type, metric_name, aggregated_value, window_size, aggregation_type)
        VALUES (?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err := stmt.Exec(metric.Timestamp, metric.ResourceType, metric.MetricName, metric.AggregatedValue, metric.WindowSize, metric.AggregationType)
		if err != nil {
			return fmt.Errorf("inserting metric: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// GetMetrics retrieves metrics with optional filtering
func (s *SQLiteStore) GetMetrics(opts GetMetricsOptions) ([]*AggregatedMetric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var conditions []string
	var args []interface{}

	if opts.ResourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, opts.ResourceType)
	}
	if opts.MetricName != "" {
		conditions = append(conditions, "metric_name = ?")
		args = append(args, opts.MetricName)
	}
	if opts.WindowSize != "" {
		conditions = append(conditions, "window_size = ?")
		args = append(args, opts.WindowSize)
	}
	if opts.AggregationType != "" {
		conditions = append(conditions, "aggregation_type = ?")
		args = append(args, opts.AggregationType)
	}
	if opts.StartTime > 0 {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, opts.StartTime)
	}
	if opts.EndTime > 0 {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, opts.EndTime)
	}

	query := "SELECT id, timestamp, resource_type, metric_name, aggregated_value, window_size, aggregation_type FROM aggregated_metrics"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp DESC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*AggregatedMetric
	for rows.Next() {
		m := &AggregatedMetric{}
		err := rows.Scan(&m.ID, &m.Timestamp, &m.ResourceType, &m.MetricName, &m.AggregatedValue, &m.WindowSize, &m.AggregationType)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// GetLatestMetrics gets the most recent metrics for a resource type and window
func (s *SQLiteStore) GetLatestMetrics(resourceType, windowSize string) ([]*AggregatedMetric, error) {
	return s.GetMetrics(GetMetricsOptions{
		ResourceType: resourceType,
		WindowSize:   windowSize,
		Limit:        10,
	})
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	return s.db.Close()
}
