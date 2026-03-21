package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type TagsRepository interface {
	Create(tag *entity.Tag) (int64, error)
	GetByID(id int64) (*entity.Tag, error)
	GetByName(name string) (*entity.Tag, error)
	GetAll() ([]*entity.Tag, error)
}
