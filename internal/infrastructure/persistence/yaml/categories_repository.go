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

func (r *CategoriesRepository) filePath() string {
	return filepath.Join(r.basePath, "categories.yaml")
}

func (r *CategoriesRepository) metaPath() string {
	return filepath.Join(r.basePath, "_meta.yaml")
}

func (r *CategoriesRepository) Create(cat *entity.Category) (int64, error) {
	if err := EnsureMetaFile(r.metaPath()); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(), "categories")
	if err != nil {
		return 0, err
	}

	cat.ID = nextID

	data, err := Read[CategoriesData](r.filePath())
	if err != nil {
		return 0, err
	}

	yamlCat := Category{}
	yamlCat.FromEntity(cat)
	data.Categories = append(data.Categories, yamlCat)

	if err := Write(r.filePath(), data); err != nil {
		return 0, err
	}

	return cat.ID, nil
}

func (r *CategoriesRepository) GetByID(id int64) (*entity.Category, error) {
	data, err := Read[CategoriesData](r.filePath())
	if err != nil {
		return nil, err
	}

	for i := range data.Categories {
		if data.Categories[i].ID == id {
			return data.Categories[i].ToEntity(), nil
		}
	}

	return nil, nil
}

func (r *CategoriesRepository) GetAll() ([]*entity.Category, error) {
	data, err := Read[CategoriesData](r.filePath())
	if err != nil {
		return nil, err
	}

	categories := make([]*entity.Category, len(data.Categories))
	for i := range data.Categories {
		categories[i] = data.Categories[i].ToEntity()
	}

	return categories, nil
}

func (r *CategoriesRepository) GetByNameOrAlias(nameOrAlias string) (*entity.Category, error) {
	nameOrAlias = entity.TrimLower(nameOrAlias)

	data, err := Read[CategoriesData](r.filePath())
	if err != nil {
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
	data, err := Read[CategoriesData](r.filePath())
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

	return Write(r.filePath(), data)
}

func (r *CategoriesRepository) Delete(id int64) error {
	data, err := Read[CategoriesData](r.filePath())
	if err != nil {
		return err
	}

	newCategories := make([]Category, 0)
	for _, cat := range data.Categories {
		if cat.ID != id {
			newCategories = append(newCategories, cat)
		}
	}

	data.Categories = newCategories
	return Write(r.filePath(), data)
}

func (r *CategoriesRepository) EnsureInitialized() error {
	path := r.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Write(path, CategoriesData{Categories: []Category{}})
	}
	return nil
}

func (r *CategoriesRepository) Port() port.CategoriesRepository {
	return r
}
