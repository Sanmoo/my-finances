package usecase

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AddCreditCardInput struct {
	Name       string
	ClosingDay int
	DueDay     int
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
	cc, err := entity.NewCreditCard(input.Name, input.ClosingDay, input.DueDay)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(cc); err != nil {
		return nil, err
	}

	return &AddCreditCardOutput{CreditCard: cc}, nil
}
