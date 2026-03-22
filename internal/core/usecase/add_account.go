package usecase

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AddAccountInput struct {
	Name string
}

type AddAccountOutput struct {
	Account *entity.Account
}

type AddAccount struct {
	repo port.AccountsRepository
}

func NewAddAccount(repo port.AccountsRepository) *AddAccount {
	return &AddAccount{repo: repo}
}

func (uc *AddAccount) Execute(input AddAccountInput) (*AddAccountOutput, error) {
	account, err := entity.NewAccount(input.Name)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(account); err != nil {
		return nil, err
	}

	return &AddAccountOutput{Account: account}, nil
}
