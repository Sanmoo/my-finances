package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type CreditCardsRepository interface {
	Create(cc *entity.CreditCard) error
	GetByName(name string) (*entity.CreditCard, error)
	GetAll() ([]*entity.CreditCard, error)
}
