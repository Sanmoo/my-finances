package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CategoriesRepository struct {
	basePath string
}

type CategoriesData struct {
	Categories []Category `yaml:"categories"`
}

func NewCategoriesRepository(basePath string) *CategoriesRepository {
	return &CategoriesRepository{basePath: basePath}
}

func (r *CategoriesRepository) filePath(accountName string) string {
	return filepath.Join(r.basePath, accountName, "categories.yaml")
}

func (r *CategoriesRepository) metaPath(accountName string) string {
	return filepath.Join(r.basePath, accountName, "_meta.yaml")
}

func (r *CategoriesRepository) getAccountName(accountID int64) (string, error) {
	accountsPath := filepath.Join(r.basePath, "accounts.yaml")
	data, err := Read[AccountsData](accountsPath)
	if err != nil {
		return "", err
	}

	for _, acc := range data.Accounts {
		if acc.ID == accountID {
			return acc.Name, nil
		}
	}
	return "", fmt.Errorf("account not found: %d", accountID)
}

func (r *CategoriesRepository) Create(cat *entity.Category) (int64, error) {
	accountName, err := r.getAccountName(cat.AccountID)
	if err != nil {
		return 0, err
	}

	metaP := r.metaPath(accountName)
	if err := EnsureMetaFile(metaP); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(metaP, "categories")
	if err != nil {
		return 0, err
	}

	cat.ID = nextID

	catPath := r.filePath(accountName)

	if err := os.MkdirAll(filepath.Dir(catPath), 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	data := &CategoriesData{}
	if _, err := os.Stat(catPath); err == nil {
		readData, err := Read[CategoriesData](catPath)
		if err != nil {
			return 0, err
		}
		data = readData
	}

	yamlCat := Category{}
	yamlCat.FromEntity(cat)
	data.Categories = append(data.Categories, yamlCat)

	if err := Write(catPath, data); err != nil {
		return 0, err
	}

	return cat.ID, nil
}

func (r *CategoriesRepository) GetByID(id int64) (*entity.Category, error) {
	accountsPath := filepath.Join(r.basePath, "accounts.yaml")
	accountsData, err := Read[AccountsData](accountsPath)
	if err != nil {
		return nil, err
	}

	for _, acc := range accountsData.Accounts {
		catPath := r.filePath(acc.Name)
		if _, err := os.Stat(catPath); os.IsNotExist(err) {
			continue
		}

		data, err := Read[CategoriesData](catPath)
		if err != nil {
			continue
		}

		for i := range data.Categories {
			if data.Categories[i].ID == id {
				return data.Categories[i].ToEntity(), nil
			}
		}
	}

	return nil, nil
}

func (r *CategoriesRepository) GetAll(accountID int64) ([]*entity.Category, error) {
	accountName, err := r.getAccountName(accountID)
	if err != nil {
		return nil, err
	}

	catPath := r.filePath(accountName)

	data, err := Read[CategoriesData](catPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entity.Category{}, nil
		}
		return nil, err
	}

	categories := make([]*entity.Category, 0, len(data.Categories))
	for i := range data.Categories {
		categories = append(categories, data.Categories[i].ToEntity())
	}

	return categories, nil
}

func (r *CategoriesRepository) GetByNameOrAlias(accountID int64, nameOrAlias string) (*entity.Category, error) {
	accountName, err := r.getAccountName(accountID)
	if err != nil {
		return nil, err
	}

	nameOrAlias = entity.TrimLower(nameOrAlias)

	catPath := r.filePath(accountName)
	data, err := Read[CategoriesData](catPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for i := range data.Categories {
		cat := data.Categories[i].ToEntity()
		if cat.Name == nameOrAlias {
			return cat, nil
		}
		if cat.Alias != nil && *cat.Alias == nameOrAlias {
			return cat, nil
		}
	}

	return nil, nil
}

func (r *CategoriesRepository) Update(cat *entity.Category) error {
	accountName, err := r.getAccountName(cat.AccountID)
	if err != nil {
		return err
	}

	catPath := r.filePath(accountName)
	data, err := Read[CategoriesData](catPath)
	if err != nil {
		return err
	}

	found := false
	for i := range data.Categories {
		if data.Categories[i].ID == cat.ID {
			data.Categories[i].FromEntity(cat)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("category not found: %d", cat.ID)
	}

	return Write(catPath, data)
}

func (r *CategoriesRepository) Delete(id int64) error {
	accountsPath := filepath.Join(r.basePath, "accounts.yaml")
	accountsData, err := Read[AccountsData](accountsPath)
	if err != nil {
		return err
	}

	for _, acc := range accountsData.Accounts {
		catPath := r.filePath(acc.Name)
		if _, err := os.Stat(catPath); os.IsNotExist(err) {
			continue
		}

		data, err := Read[CategoriesData](catPath)
		if err != nil {
			continue
		}

		found := false
		for _, cat := range data.Categories {
			if cat.ID == id {
				found = true
				break
			}
		}

		if found {
			newCategories := make([]Category, 0)
			for _, cat := range data.Categories {
				if cat.ID != id {
					newCategories = append(newCategories, cat)
				}
			}
			data.Categories = newCategories
			return Write(catPath, data)
		}
	}

	return fmt.Errorf("category not found: %d", id)
}

func (r *CategoriesRepository) EnsureInitialized() error {
	return nil
}

func (r *CategoriesRepository) Port() port.CategoriesRepository {
	return r
}
