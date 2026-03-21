package entity

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidEntryType = errors.New("entry type must be 'income' or 'expense'")
	ErrInvalidAmount    = errors.New("amount must be greater than 0")
	ErrInvalidDate      = errors.New("date cannot be empty")
	ErrEmptyCurrency    = errors.New("currency cannot be empty")
)

type EntryType string

const (
	EntryTypeIncome  EntryType = "income"
	EntryTypeExpense EntryType = "expense"
)

type Entry struct {
	ID              int64
	Type            EntryType
	Amount          float64
	Currency        string
	Description     string
	CategoryID      *int64
	CreditCardID    *int64
	RealizationDate time.Time
	PaymentDate     *time.Time
	CreatedAt       time.Time
	Tags            []string
}

func NewEntry(entryType EntryType, amount float64, currency string, realizationDate time.Time, opts ...EntryOption) (*Entry, error) {
	if entryType != EntryTypeIncome && entryType != EntryTypeExpense {
		return nil, ErrInvalidEntryType
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if strings.TrimSpace(currency) == "" {
		return nil, ErrEmptyCurrency
	}
	if realizationDate.IsZero() {
		return nil, ErrInvalidDate
	}

	entry := &Entry{
		Type:            entryType,
		Amount:          amount,
		Currency:        strings.ToUpper(currency),
		RealizationDate: realizationDate,
		CreatedAt:       time.Now().UTC(),
		Tags:            []string{},
	}

	for _, opt := range opts {
		opt(entry)
	}

	return entry, nil
}

type EntryOption func(*Entry)

func WithDescription(desc string) EntryOption {
	return func(e *Entry) {
		e.Description = desc
	}
}

func WithCategoryID(id int64) EntryOption {
	return func(e *Entry) {
		e.CategoryID = &id
	}
}

func WithCreditCard(cc *CreditCard) EntryOption {
	return func(e *Entry) {
		e.CreditCardID = &cc.ID
		paymentDate := cc.CalculatePaymentDate(e.RealizationDate)
		e.PaymentDate = &paymentDate
	}
}

func WithTags(tags []string) EntryOption {
	return func(e *Entry) {
		e.Tags = tags
	}
}

func (e *Entry) AddTag(tag string) {
	for _, t := range e.Tags {
		if t == tag {
			return
		}
	}
	e.Tags = append(e.Tags, tag)
}

func (e *Entry) IsExpense() bool {
	return e.Type == EntryTypeExpense
}

func (e *Entry) IsIncome() bool {
	return e.Type == EntryTypeIncome
}
