package port

import "github.com/Sanmoo/my-finances/internal/domain/entity"

type CreditCardsRepository interface {
	Create(cc *entity.CreditCard) (int64, error)
	GetByID(id int64) (*entity.CreditCard, error)
	GetByNamespaceID(namespaceID int64) ([]*entity.CreditCard, error)
	Update(cc *entity.CreditCard) error
	Delete(id int64) error
}
