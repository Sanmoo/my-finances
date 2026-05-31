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
	ErrEmptyDescription = errors.New("description cannot be empty")
)

type EntryType string

const (
	EntryTypeIncome  EntryType = "income"
	EntryTypeExpense EntryType = "expense"
)

type Entry struct {
	Type              EntryType
	Amount            float64
	Currency          string
	Description       string
	CategoryAlias     *string
	CreditCardName    *string
	Tags              []string
	InstallmentNumber int
	InstallmentTotal  int
	RealizationDate   time.Time
	PaymentDate       *time.Time
	CreatedAt         time.Time
}

func NewEntry(entryType EntryType, amount float64, currency string, realizationDate time.Time, opts ...EntryOption) (*Entry, error) {
	if entryType != EntryTypeIncome && entryType != EntryTypeExpense {
		return nil, ErrInvalidEntryType
	}
	if amount == 0 || (entryType == EntryTypeIncome && amount < 0) {
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

func WithCategoryAlias(alias string) EntryOption {
	return func(e *Entry) {
		e.CategoryAlias = &alias
	}
}

func WithCreditCard(cc *CreditCard) EntryOption {
	return func(e *Entry) {
		e.CreditCardName = &cc.Name
		var paymentDate time.Time
		if e.InstallmentTotal > 1 {
			paymentDate = cc.CalculateInstallmentPaymentDate(e.RealizationDate, e.InstallmentNumber)
		} else {
			paymentDate = cc.CalculatePaymentDate(e.RealizationDate)
		}
		e.PaymentDate = &paymentDate
	}
}

func WithTags(tags []string) EntryOption {
	return func(e *Entry) {
		e.Tags = tags
	}
}

func WithInstallment(number, total int) EntryOption {
	return func(e *Entry) {
		e.InstallmentNumber = number
		e.InstallmentTotal = total
	}
}

func WithPaymentDate(date time.Time) EntryOption {
	return func(e *Entry) {
		e.PaymentDate = &date
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
