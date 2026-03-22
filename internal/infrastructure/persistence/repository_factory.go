package persistence

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
)

type RepositoryFactory struct {
	basePath string
}

func NewRepositoryFactory(cfg *config.Config) *RepositoryFactory {
	return &RepositoryFactory{
		basePath: cfg.DataPath,
	}
}

func (f *RepositoryFactory) NewAccountsRepository() port.AccountsRepository {
	repo := yaml.NewAccountsRepository(f.basePath)
	repo.EnsureInitialized()
	return repo.Port()
}

func (f *RepositoryFactory) NewCategoriesRepository() port.CategoriesRepository {
	repo := yaml.NewCategoriesRepository(f.basePath)
	repo.EnsureInitialized()
	return repo.Port()
}

func (f *RepositoryFactory) NewCreditCardsRepository() port.CreditCardsRepository {
	repo := yaml.NewCreditCardsRepository(f.basePath)
	repo.EnsureInitialized()
	return repo.Port()
}

func (f *RepositoryFactory) NewEntriesRepository() port.EntriesRepository {
	repo := yaml.NewEntriesRepository(f.basePath)
	repo.EnsureInitialized()
	return repo.Port()
}

func (f *RepositoryFactory) NewTagsRepository() port.TagsRepository {
	repo := yaml.NewTagsRepository(f.basePath)
	repo.EnsureInitialized()
	return repo.Port()
}

func (f *RepositoryFactory) BasePath() string {
	return f.basePath
}
