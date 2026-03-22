package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/pkg/expr"
)

type AddEntryInput struct {
	Type           entity.EntryType
	Amount         string
	Currency       string
	Description    string
	CategoryAlias  string
	CreditCardName string
	Tags           []string
	Times          int
	Date           time.Time
	AccountName    string
}

type AddEntryOutput struct {
	Entries  []*entity.Entry
	Category *entity.Category
}

type AddEntry struct {
	entryRepo    port.EntriesRepository
	categoryRepo port.CategoriesRepository
	tagRepo      port.TagsRepository
	ccRepo       port.CreditCardsRepository
}

func NewAddEntry(
	entryRepo port.EntriesRepository,
	categoryRepo port.CategoriesRepository,
	tagRepo port.TagsRepository,
	ccRepo port.CreditCardsRepository,
) *AddEntry {
	return &AddEntry{
		entryRepo:    entryRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		ccRepo:       ccRepo,
	}
}

func (uc *AddEntry) Execute(input AddEntryInput) (*AddEntryOutput, error) {
	amount, err := expr.Parse(input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount expression: %w", err)
	}

	var creditCard *entity.CreditCard
	if input.CreditCardName != "" {
		if input.Times <= 0 {
			return nil, fmt.Errorf("--times is required when using --credit-card")
		}
		cc, err := uc.ccRepo.GetByName(input.CreditCardName)
		if err != nil {
			return nil, fmt.Errorf("failed to find credit card: %w", err)
		}
		if cc == nil {
			return nil, fmt.Errorf("credit card not found: %s", input.CreditCardName)
		}
		creditCard = cc
	} else {
		if input.Times <= 0 {
			input.Times = 1
		}
	}

	var category *entity.Category
	var categoryAlias *string
	if input.CategoryAlias != "" {
		cat, err := uc.categoryRepo.GetByAlias(input.AccountName, input.CategoryAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to find category: %w", err)
		}
		if cat == nil {
			return nil, entity.ErrCategoryNotFound
		}
		category = cat
		categoryAlias = &cat.Alias
	}

	for _, tagName := range input.Tags {
		tag, err := uc.tagRepo.GetByName(tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to validate tag %s: %w", tagName, err)
		}
		if tag == nil {
			return nil, fmt.Errorf("tag not registered: %s. Run: myfin add tag %s", tagName, tagName)
		}
	}

	entries := make([]*entity.Entry, 0, input.Times)

	for i := 0; i < input.Times; i++ {
		date := input.Date
		if i > 0 {
			date = date.AddDate(0, 1, 0)
		}

		entry, err := uc.createEntry(input, amount, date, i+1, categoryAlias, creditCard)
		if err != nil {
			return nil, err
		}

		if err := uc.entryRepo.Create(entry, input.AccountName); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return &AddEntryOutput{Entries: entries, Category: category}, nil
}

func (uc *AddEntry) createEntry(input AddEntryInput, amount float64, date time.Time, installmentNumber int, categoryAlias *string, creditCard *entity.CreditCard) (*entity.Entry, error) {
	var opts []entity.EntryOption

	description := strings.TrimSpace(input.Description)
	if description == "" {
		return nil, entity.ErrEmptyDescription
	}
	opts = append(opts, entity.WithDescription(description))

	if categoryAlias != nil {
		opts = append(opts, entity.WithCategoryAlias(*categoryAlias))
	}

	if creditCard != nil {
		opts = append(opts, entity.WithCreditCard(creditCard))
	}

	if len(input.Tags) > 0 {
		opts = append(opts, entity.WithTags(input.Tags))
	}

	if input.Times > 1 {
		opts = append(opts, entity.WithInstallment(installmentNumber, input.Times))
	}

	entry, err := entity.NewEntry(input.Type, amount, input.Currency, date, opts...)
	if err != nil {
		return nil, err
	}

	return entry, nil
}
