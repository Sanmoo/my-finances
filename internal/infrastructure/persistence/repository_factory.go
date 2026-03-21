package persistence

import (
	"database/sql"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/sqlite"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
)

type RepositoryFactory struct {
	driver   string
	basePath string
	db       *sql.DB
}

func NewRepositoryFactory(cfg *config.Config, db *sql.DB) *RepositoryFactory {
	return &RepositoryFactory{
		driver:   cfg.StorageDriver,
		basePath: cfg.DataPath,
		db:       db,
	}
}

func (f *RepositoryFactory) NewAccountsRepository() port.AccountsRepository {
	switch f.driver {
	case config.DriverYAML:
		repo := yaml.NewAccountsRepository(f.basePath)
		repo.EnsureInitialized()
		return repo.Port()
	case config.DriverSQLite:
		fallthrough
	default:
		return sqlite.NewAccountsRepository(&sqlite.DB{DB: f.db})
	}
}

func (f *RepositoryFactory) NewCategoriesRepository() port.CategoriesRepository {
	switch f.driver {
	case config.DriverYAML:
		repo := yaml.NewCategoriesRepository(f.basePath)
		repo.EnsureInitialized()
		return repo.Port()
	case config.DriverSQLite:
		fallthrough
	default:
		return sqlite.NewCategoriesRepository(&sqlite.DB{DB: f.db})
	}
}

func (f *RepositoryFactory) NewCreditCardsRepository() port.CreditCardsRepository {
	switch f.driver {
	case config.DriverYAML:
		repo := yaml.NewCreditCardsRepository(f.basePath)
		repo.EnsureInitialized()
		return repo.Port()
	case config.DriverSQLite:
		fallthrough
	default:
		return sqlite.NewCreditCardsRepository(&sqlite.DB{DB: f.db})
	}
}

func (f *RepositoryFactory) NewEntriesRepository() port.EntriesRepository {
	switch f.driver {
	case config.DriverYAML:
		repo := yaml.NewEntriesRepository(f.basePath)
		repo.EnsureInitialized()
		return repo.Port()
	case config.DriverSQLite:
		fallthrough
	default:
		return sqlite.NewEntriesRepository(&sqlite.DB{DB: f.db})
	}
}

func (f *RepositoryFactory) Driver() string {
	return f.driver
}

func (f *RepositoryFactory) BasePath() string {
	return f.basePath
}
