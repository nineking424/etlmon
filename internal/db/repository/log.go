package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
)

// LogRepository handles log entry data access
type LogRepository struct {
	db         *sql.DB
	stmtInsert *sql.Stmt
	stmtGet    *sql.Stmt
	stmtTrim   *sql.Stmt
}

// NewLogRepository creates a new LogRepository with prepared statements
func NewLogRepository(db *sql.DB) *LogRepository {
	r := &LogRepository{db: db}

	var err error
	r.stmtInsert, err = db.Prepare(`
		INSERT INTO log_lines (log_name, log_path, line, created_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare log insert statement: %v", err))
	}

	r.stmtGet, err = db.Prepare(`
		SELECT id, log_name, log_path, line, created_at
		FROM log_lines
		WHERE log_name = ?
		ORDER BY id DESC
		LIMIT ?
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare log select statement: %v", err))
	}

	r.stmtTrim, err = db.Prepare(`
		DELETE FROM log_lines
		WHERE log_name = ? AND id NOT IN (
			SELECT id FROM log_lines WHERE log_name = ? ORDER BY id DESC LIMIT ?
		)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare log trim statement: %v", err))
	}

	return r
}

// SaveLogEntry inserts a new log entry
func (r *LogRepository) SaveLogEntry(ctx context.Context, entry *models.LogEntry) error {
	_, err := r.stmtInsert.ExecContext(ctx,
		entry.LogName,
		entry.LogPath,
		entry.Line,
		entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save log entry: %w", err)
	}
	return nil
}

// GetLogEntries retrieves recent log entries for a specific log
func (r *LogRepository) GetLogEntries(ctx context.Context, logName string, limit int) ([]*models.LogEntry, error) {
	rows, err := r.stmtGet.QueryContext(ctx, logName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer rows.Close()

	var result []*models.LogEntry
	for rows.Next() {
		e := &models.LogEntry{}
		err := rows.Scan(&e.ID, &e.LogName, &e.LogPath, &e.Line, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry row: %w", err)
		}
		result = append(result, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating log entry rows: %w", err)
	}

	// Reverse to get chronological order (we queried DESC for LIMIT)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// GetAllLogEntries retrieves recent log entries across all logs
func (r *LogRepository) GetAllLogEntries(ctx context.Context, limit int) ([]*models.LogEntry, error) {
	query := `
		SELECT id, log_name, log_path, line, created_at
		FROM log_lines
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query all log entries: %w", err)
	}
	defer rows.Close()

	var result []*models.LogEntry
	for rows.Next() {
		e := &models.LogEntry{}
		err := rows.Scan(&e.ID, &e.LogName, &e.LogPath, &e.Line, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry row: %w", err)
		}
		result = append(result, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating log entry rows: %w", err)
	}

	// Reverse to chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// ListAll returns all log entries (most recent 200)
func (r *LogRepository) ListAll() ([]models.LogEntry, error) {
	results, err := r.GetAllLogEntries(context.Background(), 200)
	if err != nil {
		return nil, err
	}
	var list []models.LogEntry
	for _, item := range results {
		list = append(list, *item)
	}
	return list, nil
}

// TrimOldEntries removes old entries beyond maxLines for a specific log
func (r *LogRepository) TrimOldEntries(ctx context.Context, logName string, maxLines int) error {
	_, err := r.stmtTrim.ExecContext(ctx, logName, logName, maxLines)
	if err != nil {
		return fmt.Errorf("failed to trim log entries: %w", err)
	}
	return nil
}

// Close closes prepared statements
func (r *LogRepository) Close() error {
	var errs []error
	if err := r.stmtInsert.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.stmtGet.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.stmtTrim.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close statements: %v", errs)
	}
	return nil
}
