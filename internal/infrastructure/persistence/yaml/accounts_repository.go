package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AccountsRepository struct {
	basePath string
}

type AccountsData struct {
	Accounts []Account `yaml:"accounts"`
}

func NewAccountsRepository(basePath string) *AccountsRepository {
	return &AccountsRepository{basePath: basePath}
}

func (r *AccountsRepository) filePath() string {
	return filepath.Join(r.basePath, "accounts.yaml")
}

func (r *AccountsRepository) metaPath() string {
	return filepath.Join(r.basePath, "_meta.yaml")
}

func (r *AccountsRepository) Create(acc *entity.Account) (int64, error) {
	if err := EnsureMetaFile(r.metaPath()); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(), "accounts")
	if err != nil {
		return 0, err
	}

	acc.ID = nextID

	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return 0, err
	}

	yamlAcc := Account{
		ID:   acc.ID,
		Name: acc.Name,
	}
	data.Accounts = append(data.Accounts, yamlAcc)

	if err := Write(r.filePath(), data); err != nil {
		return 0, err
	}

	return acc.ID, nil
}

func (r *AccountsRepository) GetByID(id int64) (*entity.Account, error) {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return nil, err
	}

	for i := range data.Accounts {
		if data.Accounts[i].ID == id {
			return &entity.Account{
				ID:   data.Accounts[i].ID,
				Name: data.Accounts[i].Name,
			}, nil
		}
	}

	return nil, nil
}

func (r *AccountsRepository) GetByName(name string) (*entity.Account, error) {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return nil, err
	}

	for i := range data.Accounts {
		if data.Accounts[i].Name == name {
			return &entity.Account{
				ID:   data.Accounts[i].ID,
				Name: data.Accounts[i].Name,
			}, nil
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
	for i := range data.Accounts {
		accounts[i] = &entity.Account{
			ID:   data.Accounts[i].ID,
			Name: data.Accounts[i].Name,
		}
	}

	return accounts, nil
}

func (r *AccountsRepository) Update(acc *entity.Account) error {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return err
	}

	found := false
	for i := range data.Accounts {
		if data.Accounts[i].ID == acc.ID {
			data.Accounts[i].Name = acc.Name
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("account not found: %d", acc.ID)
	}

	return Write(r.filePath(), data)
}

func (r *AccountsRepository) Delete(id int64) error {
	data, err := Read[AccountsData](r.filePath())
	if err != nil {
		return err
	}

	newAccounts := make([]Account, 0)
	for _, acc := range data.Accounts {
		if acc.ID != id {
			newAccounts = append(newAccounts, acc)
		}
	}

	data.Accounts = newAccounts
	return Write(r.filePath(), data)
}

func (r *AccountsRepository) EnsureInitialized() error {
	path := r.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Write(path, AccountsData{Accounts: []Account{}})
	}
	return nil
}

func (r *AccountsRepository) Port() port.AccountsRepository {
	return r
}
