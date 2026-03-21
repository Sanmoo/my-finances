package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type CategoriesRepository interface {
	Create(cat *entity.Category) (int64, error)
	GetByID(id int64) (*entity.Category, error)
	GetByNamespaceID(namespaceID int64) ([]*entity.Category, error)
	GetByNameOrAlias(namespaceID int64, nameOrAlias string) (*entity.Category, error)
	Update(cat *entity.Category) error
	Delete(id int64) error
}
