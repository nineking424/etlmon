package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/etlmon/etlmon/internal/db/schema"
)

// DB wraps sql.DB with application-specific functionality
type DB struct {
	db *sql.DB
}

// NewDB creates a new database connection with WAL mode and runs migrations
func NewDB(dbPath string) (*DB, error) {
	// Open with WAL mode and optimized settings
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL", dbPath)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := schema.RunMigrations(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{db: sqlDB}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// WithTx executes a function within a transaction
// Commits on success, rolls back on error
func (d *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute function
	err = fn(tx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Compact runs VACUUM to reclaim space and optimize database
func (d *DB) Compact(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("failed to compact database: %w", err)
	}
	return nil
}

// GetDB returns the underlying *sql.DB for use by repositories
func (d *DB) GetDB() *sql.DB {
	return d.db
}
