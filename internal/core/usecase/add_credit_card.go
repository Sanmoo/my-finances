package usecase

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AddCreditCardInput struct {
	NamespaceID int64
	Name        string
	ClosingDay  int
	DueDay      int
}

type AddCreditCardOutput struct {
	CreditCard *entity.CreditCard
}

type AddCreditCard struct {
	repo port.CreditCardsRepository
}

func NewAddCreditCard(repo port.CreditCardsRepository) *AddCreditCard {
	return &AddCreditCard{repo: repo}
}

func (uc *AddCreditCard) Execute(input AddCreditCardInput) (*AddCreditCardOutput, error) {
	cc, err := entity.NewCreditCard(input.NamespaceID, input.Name, input.ClosingDay, input.DueDay)
	if err != nil {
		return nil, err
	}

	id, err := uc.repo.Create(cc)
	if err != nil {
		return nil, err
	}

	cc.ID = id
	return &AddCreditCardOutput{CreditCard: cc}, nil
}
