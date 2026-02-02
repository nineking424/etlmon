package repository

import "database/sql"

// Repository aggregates all sub-repositories
type Repository struct {
	FS    *FSRepository
	Paths *PathsRepository
}

// NewRepository creates a new Repository with all sub-repositories initialized
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		FS:    NewFSRepository(db),
		Paths: NewPathsRepository(db),
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

	if len(errs) > 0 {
		return errs[0] // Return first error
	}
	return nil
}
