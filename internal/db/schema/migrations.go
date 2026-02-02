package schema

import (
	"database/sql"
	_ "embed"
)

//go:embed 001_initial.sql
var migration001 string

// RunMigrations executes all database migrations in order.
// Uses CREATE TABLE IF NOT EXISTS for idempotency.
func RunMigrations(db *sql.DB) error {
	// Execute initial migration
	_, err := db.Exec(migration001)
	return err
}
