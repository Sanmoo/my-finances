package yaml

import (
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CategoriesRepository struct {
	basePath string
}

func NewCategoriesRepository(basePath string) *CategoriesRepository {
	return &CategoriesRepository{basePath: basePath}
}

func (r *CategoriesRepository) filePath(accountName string) string {
	return filepath.Join(r.basePath, accountName, "categories.yaml")
}

func (r *CategoriesRepository) Create(cat *entity.Category, accountName string) error {
	catPath := r.filePath(accountName)

	if err := os.MkdirAll(filepath.Dir(catPath), 0755); err != nil {
		return err
	}

	data := &CategoriesData{}
	if _, err := os.Stat(catPath); err == nil {
		readData, err := Read[CategoriesData](catPath)
		if err != nil {
			return err
		}
		data = readData
	}

	yamlCat := Category{}
	yamlCat.FromEntity(cat)
	data.Categories = append(data.Categories, yamlCat)

	return Write(catPath, data)
}

func (r *CategoriesRepository) GetByAlias(accountName string, alias string) (*entity.Category, error) {
	alias = entity.TrimLower(alias)

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
		if cat.Alias == alias {
			return cat, nil
		}
	}

	return nil, nil
}

func (r *CategoriesRepository) GetAll(accountName string) ([]*entity.Category, error) {
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

func (r *CategoriesRepository) EnsureInitialized() error {
	return nil
}

func (r *CategoriesRepository) Port() port.CategoriesRepository {
	return r
}
