package persistence

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
)

type MigrationManager struct {
	db             *sql.DB
	migrationsPath string
}

func NewMigrationManager(db *sql.DB, migrationsPath string) *MigrationManager {
	return &MigrationManager{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

func (m *MigrationManager) Up() error {
	dbURL := "file://myfin.db"

	mi, err := migrate.New(m.migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer mi.Close()

	if err := mi.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (m *MigrationManager) Down(steps int) error {
	dbURL := "file://myfin.db"

	mi, err := migrate.New(m.migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer mi.Close()

	if err := mi.Steps(-steps); err != nil {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}

	return nil
}

func (m *MigrationManager) Version() (uint, bool, error) {
	dbURL := "file://myfin.db"

	mi, err := migrate.New(m.migrationsPath, dbURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer mi.Close()

	version, dirty, err := mi.Version()
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	mm := NewMigrationManager(db, migrationsPath)
	return mm.Up()
}
