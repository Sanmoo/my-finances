package entity

import "errors"

var (
	ErrEmptyAccountName = errors.New("account name cannot be empty")
)

type Account struct {
	Name string
}

func NewAccount(name string) (*Account, error) {
	name = TrimLower(name)
	if name == "" {
		return nil, ErrEmptyAccountName
	}

	return &Account{
		Name: name,
	}, nil
}
