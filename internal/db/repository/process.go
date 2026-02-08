package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
)

// ProcessRepository handles process info data access
type ProcessRepository struct {
	db         *sql.DB
	stmtInsert *sql.Stmt
	stmtGetAll *sql.Stmt
	stmtClear  *sql.Stmt
}

// NewProcessRepository creates a new ProcessRepository with prepared statements
func NewProcessRepository(db *sql.DB) *ProcessRepository {
	r := &ProcessRepository{db: db}

	var err error
	r.stmtInsert, err = db.Prepare(`
		INSERT OR REPLACE INTO process_stats
		(pid, name, user, cpu_percent, mem_rss, status, elapsed, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare process insert statement: %v", err))
	}

	r.stmtGetAll, err = db.Prepare(`
		SELECT pid, name, user, cpu_percent, mem_rss, status, elapsed, collected_at
		FROM process_stats
		ORDER BY cpu_percent DESC
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare process select statement: %v", err))
	}

	r.stmtClear, err = db.Prepare(`DELETE FROM process_stats`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare process clear statement: %v", err))
	}

	return r
}

// SaveProcessInfo inserts or updates a process record
func (r *ProcessRepository) SaveProcessInfo(ctx context.Context, info *models.ProcessInfo) error {
	_, err := r.stmtInsert.ExecContext(ctx,
		info.PID,
		info.Name,
		info.User,
		info.CPUPercent,
		info.MemRSS,
		info.Status,
		info.Elapsed,
		info.CollectedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save process info: %w", err)
	}
	return nil
}

// GetLatestProcessInfo retrieves all process records ordered by CPU usage
func (r *ProcessRepository) GetLatestProcessInfo(ctx context.Context) ([]*models.ProcessInfo, error) {
	rows, err := r.stmtGetAll.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query process info: %w", err)
	}
	defer rows.Close()

	var result []*models.ProcessInfo
	for rows.Next() {
		p := &models.ProcessInfo{}
		err := rows.Scan(
			&p.PID,
			&p.Name,
			&p.User,
			&p.CPUPercent,
			&p.MemRSS,
			&p.Status,
			&p.Elapsed,
			&p.CollectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan process info row: %w", err)
		}
		result = append(result, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating process info rows: %w", err)
	}
	return result, nil
}

// ListAll returns all process records
func (r *ProcessRepository) ListAll() ([]models.ProcessInfo, error) {
	results, err := r.GetLatestProcessInfo(context.Background())
	if err != nil {
		return nil, err
	}
	var list []models.ProcessInfo
	for _, item := range results {
		list = append(list, *item)
	}
	return list, nil
}

// ClearAll removes all process records (called before each fresh collection)
func (r *ProcessRepository) ClearAll(ctx context.Context) error {
	_, err := r.stmtClear.ExecContext(ctx)
	return err
}

// Close closes prepared statements
func (r *ProcessRepository) Close() error {
	var errs []error
	if err := r.stmtInsert.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.stmtGetAll.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.stmtClear.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close statements: %v", errs)
	}
	return nil
}
