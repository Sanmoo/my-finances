package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type CategoriesRepository interface {
	Create(cat *entity.Category, accountName string) error
	GetByAlias(accountName string, alias string) (*entity.Category, error)
	GetAll(accountName string) ([]*entity.Category, error)
}
