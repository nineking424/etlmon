package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
)

// PathsRepository handles path statistics data access
type PathsRepository struct {
	db          *sql.DB
	stmtInsert  *sql.Stmt
	stmtGetAll  *sql.Stmt
}

// NewPathsRepository creates a new PathsRepository with prepared statements
func NewPathsRepository(db *sql.DB) *PathsRepository {
	r := &PathsRepository{db: db}

	var err error
	r.stmtInsert, err = db.Prepare(`
		INSERT OR REPLACE INTO path_stats
		(path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare insert statement: %v", err))
	}

	r.stmtGetAll, err = db.Prepare(`
		SELECT path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at
		FROM path_stats
		ORDER BY path
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare select statement: %v", err))
	}

	return r
}

// Save inserts or updates path statistics record
func (r *PathsRepository) Save(ctx context.Context, stats *models.PathStats) error {
	_, err := r.stmtInsert.ExecContext(ctx,
		stats.Path,
		stats.FileCount,
		stats.DirCount,
		stats.ScanDurationMs,
		stats.Status,
		stats.ErrorMessage,
		stats.CollectedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save path stats: %w", err)
	}
	return nil
}

// GetAll retrieves all path statistics records ordered by path
func (r *PathsRepository) GetAll(ctx context.Context) ([]*models.PathStats, error) {
	rows, err := r.stmtGetAll.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query path stats: %w", err)
	}
	defer rows.Close()

	var result []*models.PathStats
	for rows.Next() {
		s := &models.PathStats{}
		err := rows.Scan(
			&s.Path,
			&s.FileCount,
			&s.DirCount,
			&s.ScanDurationMs,
			&s.Status,
			&s.ErrorMessage,
			&s.CollectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan path stats row: %w", err)
		}
		result = append(result, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating path stats rows: %w", err)
	}

	return result, nil
}

// ListAll returns all path statistics records (alias for GetAll with empty context)
func (r *PathsRepository) ListAll() ([]models.PathStats, error) {
	query := `
		SELECT path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at
		FROM path_stats
		ORDER BY path
	`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query path stats: %w", err)
	}
	defer rows.Close()

	var results []models.PathStats
	for rows.Next() {
		var ps models.PathStats
		var errMsg sql.NullString
		if err := rows.Scan(&ps.Path, &ps.FileCount, &ps.DirCount,
			&ps.ScanDurationMs, &ps.Status, &errMsg, &ps.CollectedAt); err != nil {
			return nil, fmt.Errorf("failed to scan path stats row: %w", err)
		}
		if errMsg.Valid {
			ps.ErrorMessage = errMsg.String
		}
		results = append(results, ps)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating path stats rows: %w", err)
	}

	return results, nil
}

// ListWithPagination returns path statistics with limit and offset
func (r *PathsRepository) ListWithPagination(limit, offset int) ([]models.PathStats, error) {
	query := `
		SELECT path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at
		FROM path_stats
		ORDER BY path
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(context.Background(), query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query path stats with pagination: %w", err)
	}
	defer rows.Close()

	var results []models.PathStats
	for rows.Next() {
		var ps models.PathStats
		var errMsg sql.NullString
		if err := rows.Scan(&ps.Path, &ps.FileCount, &ps.DirCount,
			&ps.ScanDurationMs, &ps.Status, &errMsg, &ps.CollectedAt); err != nil {
			return nil, fmt.Errorf("failed to scan path stats row: %w", err)
		}
		if errMsg.Valid {
			ps.ErrorMessage = errMsg.String
		}
		results = append(results, ps)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating path stats rows: %w", err)
	}

	return results, nil
}

// Count returns the total number of path records
func (r *PathsRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM path_stats").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count path stats: %w", err)
	}
	return count, nil
}

// SavePathStats is an alias for Save to match collector interface
func (r *PathsRepository) SavePathStats(ctx context.Context, stats *models.PathStats) error {
	return r.Save(ctx, stats)
}

// GetLatestPathStats is an alias for GetAll to match collector interface
func (r *PathsRepository) GetLatestPathStats(ctx context.Context) ([]*models.PathStats, error) {
	return r.GetAll(ctx)
}

// GetPathStats retrieves statistics for a specific path
func (r *PathsRepository) GetPathStats(ctx context.Context, path string) (*models.PathStats, error) {
	query := `
		SELECT path, file_count, dir_count, scan_duration_ms, status, error_message, collected_at
		FROM path_stats
		WHERE path = ?
	`

	var stats models.PathStats
	var errMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&stats.Path,
		&stats.FileCount,
		&stats.DirCount,
		&stats.ScanDurationMs,
		&stats.Status,
		&errMsg,
		&stats.CollectedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get path stats for %s: %w", path, err)
	}

	if errMsg.Valid {
		stats.ErrorMessage = errMsg.String
	}

	return &stats, nil
}

// Close closes prepared statements
func (r *PathsRepository) Close() error {
	var errs []error

	if err := r.stmtInsert.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.stmtGetAll.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close statements: %v", errs)
	}
	return nil
}
