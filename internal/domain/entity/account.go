package entity

import "errors"

var (
	ErrEmptyAccountName = errors.New("account name cannot be empty")
)

type Account struct {
	ID          int64
	NamespaceID int64
	Name        string
}

func NewAccount(namespaceID int64, name string) (*Account, error) {
	name = trimLower(name)
	if name == "" {
		return nil, ErrEmptyAccountName
	}

	return &Account{
		NamespaceID: namespaceID,
		Name:        name,
	}, nil
}
