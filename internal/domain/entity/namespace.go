package entity

import (
	"errors"
	"time"
)

var (
	ErrEmptyNamespaceName = errors.New("namespace name cannot be empty")
	ErrInvalidCurrency    = errors.New("currency code must be 3 characters")
)

type Namespace struct {
	ID                  int64
	Name                string
	DefaultCreditCardID *int64
	DefaultCurrency     string
}

func NewNamespace(name string, opts ...NamespaceOption) (*Namespace, error) {
	if name = trimLower(name); name == "" {
		return nil, ErrEmptyNamespaceName
	}

	ns := &Namespace{
		Name:            name,
		DefaultCurrency: "BRL",
	}

	for _, opt := range opts {
		opt(ns)
	}

	if len(ns.DefaultCurrency) != 3 {
		return nil, ErrInvalidCurrency
	}

	return ns, nil
}

type NamespaceOption func(*Namespace)

func WithDefaultCreditCard(ccID int64) NamespaceOption {
	return func(ns *Namespace) {
		ns.DefaultCreditCardID = &ccID
	}
}

func WithDefaultCurrency(currency string) NamespaceOption {
	return func(ns *Namespace) {
		ns.DefaultCurrency = currency
	}
}

func (ns *Namespace) CreatedAt() time.Time {
	return time.Now().UTC()
}

func trimLower(s string) string {
	return trim(s)
}

func trim(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\n' {
			start := i
			for j := len(s) - 1; j >= start; j-- {
				if s[j] != ' ' && s[j] != '\t' && s[j] != '\n' {
					return toLower(s[start : j+1])
				}
			}
		}
	}
	return ""
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
