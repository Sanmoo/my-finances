package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type AccountsRepository interface {
	Create(acc *entity.Account) (int64, error)
	GetByID(id int64) (*entity.Account, error)
	GetAll() ([]*entity.Account, error)
	Update(acc *entity.Account) error
	Delete(id int64) error
}
