package usecase

import (
	"fmt"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/pkg/expr"
)

type AddEntryInput struct {
	Type        entity.EntryType
	Amount      string
	Currency    string
	Description string
	CategoryID  *int64
	CreditCard  *entity.CreditCard
	Tags        []string
	Times       int
	Date        time.Time
	AccountID   int64
}

type AddEntryOutput struct {
	Entries []*entity.Entry
}

type AddEntry struct {
	entryRepo    port.EntriesRepository
	categoryRepo port.CategoriesRepository
	ccRepo       port.CreditCardsRepository
}

func NewAddEntry(
	entryRepo port.EntriesRepository,
	categoryRepo port.CategoriesRepository,
	ccRepo port.CreditCardsRepository,
) *AddEntry {
	return &AddEntry{
		entryRepo:    entryRepo,
		categoryRepo: categoryRepo,
		ccRepo:       ccRepo,
	}
}

func (uc *AddEntry) Execute(input AddEntryInput) (*AddEntryOutput, error) {
	amount, err := expr.Parse(input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount expression: %w", err)
	}

	if input.Times <= 0 {
		input.Times = 1
	}

	entries := make([]*entity.Entry, 0, input.Times)
	var parentID *int64

	for i := 0; i < input.Times; i++ {
		date := input.Date
		if i > 0 {
			date = date.AddDate(0, 1, 0)
		}

		entry, err := uc.createEntry(input, amount, date, i+1, parentID)
		if err != nil {
			return nil, err
		}

		id, err := uc.entryRepo.Create(entry)
		if err != nil {
			return nil, err
		}

		entry.ID = id
		entries = append(entries, entry)

		if parentID == nil {
			parentID = &entry.ID
		}
	}

	return &AddEntryOutput{Entries: entries}, nil
}

func (uc *AddEntry) createEntry(input AddEntryInput, amount float64, date time.Time, installment int, parentID *int64) (*entity.Entry, error) {
	var opts []entity.EntryOption

	if input.Description != "" {
		opts = append(opts, entity.WithDescription(input.Description))
	}

	if input.CategoryID != nil {
		opts = append(opts, entity.WithCategoryID(*input.CategoryID))
	}

	if input.CreditCard != nil {
		opts = append(opts, entity.WithCreditCard(input.CreditCard))
	}

	if len(input.Tags) > 0 {
		opts = append(opts, entity.WithTags(input.Tags))
	}

	opts = append(opts, entity.WithAccountID(input.AccountID))

	if installment > 0 || parentID != nil {
		opts = append(opts, entity.WithInstallment(installment, parentID))
	}

	entry, err := entity.NewEntry(input.Type, amount, input.Currency, date, opts...)
	if err != nil {
		return nil, err
	}

	return entry, nil
}
