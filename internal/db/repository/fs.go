package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/etlmon/etlmon/pkg/models"
)

// FSRepository handles filesystem usage data access
type FSRepository struct {
	db          *sql.DB
	stmtInsert  *sql.Stmt
	stmtGetAll  *sql.Stmt
}

// NewFSRepository creates a new FSRepository with prepared statements
func NewFSRepository(db *sql.DB) *FSRepository {
	r := &FSRepository{db: db}

	var err error
	r.stmtInsert, err = db.Prepare(`
		INSERT OR REPLACE INTO filesystem_usage
		(mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare insert statement: %v", err))
	}

	r.stmtGetAll, err = db.Prepare(`
		SELECT mount_point, total_bytes, used_bytes, avail_bytes, used_percent, collected_at
		FROM filesystem_usage
		ORDER BY mount_point
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to prepare select statement: %v", err))
	}

	return r
}

// Save inserts or updates filesystem usage record
func (r *FSRepository) Save(ctx context.Context, usage *models.FilesystemUsage) error {
	_, err := r.stmtInsert.ExecContext(ctx,
		usage.MountPoint,
		usage.TotalBytes,
		usage.UsedBytes,
		usage.AvailBytes,
		usage.UsedPercent,
		usage.CollectedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save filesystem usage: %w", err)
	}
	return nil
}

// GetLatest retrieves all filesystem usage records ordered by mount point
func (r *FSRepository) GetLatest(ctx context.Context) ([]*models.FilesystemUsage, error) {
	rows, err := r.stmtGetAll.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query filesystem usage: %w", err)
	}
	defer rows.Close()

	var result []*models.FilesystemUsage
	for rows.Next() {
		u := &models.FilesystemUsage{}
		err := rows.Scan(
			&u.MountPoint,
			&u.TotalBytes,
			&u.UsedBytes,
			&u.AvailBytes,
			&u.UsedPercent,
			&u.CollectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filesystem usage row: %w", err)
		}
		result = append(result, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filesystem usage rows: %w", err)
	}

	return result, nil
}

// ListAll returns all filesystem usage records (alias for GetLatest with empty context)
func (r *FSRepository) ListAll() ([]models.FilesystemUsage, error) {
	results, err := r.GetLatest(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert from []*models.FilesystemUsage to []models.FilesystemUsage
	var list []models.FilesystemUsage
	for _, item := range results {
		list = append(list, *item)
	}
	return list, nil
}

// SaveFilesystemUsage is an alias for Save to match collector interface
func (r *FSRepository) SaveFilesystemUsage(ctx context.Context, usage *models.FilesystemUsage) error {
	return r.Save(ctx, usage)
}

// GetLatestFilesystemUsage is an alias for GetLatest to match collector interface
func (r *FSRepository) GetLatestFilesystemUsage(ctx context.Context) ([]*models.FilesystemUsage, error) {
	return r.GetLatest(ctx)
}

// Close closes prepared statements
func (r *FSRepository) Close() error {
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
