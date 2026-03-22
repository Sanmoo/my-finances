package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type AccountsRepository interface {
	Create(acc *entity.Account) error
	GetByName(name string) (*entity.Account, error)
	GetAll() ([]*entity.Account, error)
}
