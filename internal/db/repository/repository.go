package repository

import "database/sql"

// Repository aggregates all sub-repositories
type Repository struct {
	FS      *FSRepository
	Paths   *PathsRepository
	Process *ProcessRepository
	Log     *LogRepository
}

// NewRepository creates a new Repository with all sub-repositories initialized
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		FS:      NewFSRepository(db),
		Paths:   NewPathsRepository(db),
		Process: NewProcessRepository(db),
		Log:     NewLogRepository(db),
	}
}

// Close closes all sub-repositories and their prepared statements
func (r *Repository) Close() error {
	var errs []error

	if err := r.FS.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.Paths.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.Process.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := r.Log.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs[0] // Return first error
	}
	return nil
}
