package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/sqlite"
)

type Manager struct {
	cfgLoader *config.Loader
}

func NewManager(cfgLoader *config.Loader) *Manager {
	return &Manager{cfgLoader: cfgLoader}
}

func (m *Manager) GetDatabasesPath() (string, error) {
	cfg, err := m.cfgLoader.Load()
	if err != nil {
		return "", err
	}

	baseDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if cfg.DatabasesPath != "" {
		return cfg.DatabasesPath, nil
	}

	return filepath.Join(baseDir, ".myfin", "databases"), nil
}

func (m *Manager) GetDatabasePath(name string) (string, error) {
	if name == "" {
		name = "default"
	}

	databasesPath, err := m.GetDatabasesPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(databasesPath, name+".db"), nil
}

func (m *Manager) GetDatabase(name string) (*sqlite.DB, error) {
	if name == "" {
		name = "default"
	}

	dbPath, err := m.GetDatabasePath(name)
	if err != nil {
		return nil, err
	}

	databasesPath, err := m.GetDatabasesPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(databasesPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create databases directory: %w", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := m.createDatabase(dbPath); err != nil {
			return nil, err
		}
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func (m *Manager) createDatabase(dbPath string) error {
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Close()

	absMigrationsPath, _ := filepath.Abs("migrations")
	mm := persistence.NewMigrationManager(db.DB, "file://"+absMigrationsPath)
	if err := mm.Up(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (m *Manager) ListDatabases() ([]string, error) {
	databasesPath, err := m.GetDatabasesPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(databasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read databases directory: %w", err)
	}

	var databases []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			name := entry.Name()[:len(entry.Name())-3]
			databases = append(databases, name)
		}
	}

	return databases, nil
}

func (m *Manager) CreateDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	dbPath, err := m.GetDatabasePath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database '%s' already exists", name)
	}

	return m.createDatabase(dbPath)
}

func (m *Manager) DeleteDatabase(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default database")
	}

	dbPath, err := m.GetDatabasePath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database '%s' does not exist", name)
	}

	return os.Remove(dbPath)
}
