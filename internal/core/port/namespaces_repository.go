package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type NamespacesRepository interface {
	Create(ns *entity.Namespace) (int64, error)
	GetByID(id int64) (*entity.Namespace, error)
	GetByName(name string) (*entity.Namespace, error)
	GetAll() ([]*entity.Namespace, error)
	Update(ns *entity.Namespace) error
	Delete(id int64) error
}
