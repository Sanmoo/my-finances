package yaml

import (
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AccountsRepository struct {
	basePath string
}

func NewAccountsRepository(basePath string) *AccountsRepository {
	return &AccountsRepository{basePath: basePath}
}

func (r *AccountsRepository) filePath() string {
	return filepath.Join(r.basePath, "accounts.yaml")
}

func (r *AccountsRepository) Create(acc *entity.Account) error {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return err
	}

	for _, name := range data.Accounts {
		if name == acc.Name {
			return nil
		}
	}

	data.Accounts = append(data.Accounts, acc.Name)

	return Write(r.filePath(), data)
}

func (r *AccountsRepository) GetByName(name string) (*entity.Account, error) {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return nil, err
	}

	for _, accountName := range data.Accounts {
		if accountName == name {
			return &entity.Account{Name: name}, nil
		}
	}

	return nil, nil
}

func (r *AccountsRepository) GetAll() ([]*entity.Account, error) {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return nil, err
	}

	accounts := make([]*entity.Account, len(data.Accounts))
	for i, name := range data.Accounts {
		accounts[i] = &entity.Account{Name: name}
	}

	return accounts, nil
}

func (r *AccountsRepository) EnsureInitialized() error {
	path := r.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Write(path, AccountsData{Accounts: []string{}})
	}
	return nil
}

func (r *AccountsRepository) Port() port.AccountsRepository {
	return r
}
