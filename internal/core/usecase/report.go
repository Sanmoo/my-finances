package usecase

import (
	"fmt"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type ReportInput struct {
	Format                  string
	From                    *time.Time
	To                      *time.Time
	FilterTags              []string
	FilterCategories        []string
	AccountName             string
	FilterByRealizationDate bool
}

type ReportOutput struct {
	Entries      []*EntryWithCategory
	Accounts     []*entity.Account
	TotalIncome  float64
	TotalExpense float64
}

type EntryWithCategory struct {
	Type              string
	Amount            float64
	Currency          string
	Description       string
	CategoryAlias     *string
	CategoryName      string
	CreditCardName    *string
	AccountName       string
	InstallmentNumber int
	InstallmentTotal  int
	RealizationDate   time.Time
	PaymentDate       *time.Time
	Tags              []string
}

type Report struct {
	entryRepo    port.EntriesRepository
	categoryRepo port.CategoriesRepository
	accountRepo  port.AccountsRepository
}

func NewReport(entryRepo port.EntriesRepository, categoryRepo port.CategoriesRepository, accountRepo port.AccountsRepository) *Report {
	return &Report{
		entryRepo:    entryRepo,
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
	}
}

func (uc *Report) Execute(input ReportInput) (*ReportOutput, error) {
	filters := &port.EntryFilters{
		FromDate:                input.From,
		ToDate:                  input.To,
		AccountName:             input.AccountName,
		FilterByRealizationDate: input.FilterByRealizationDate,
	}

	if len(input.FilterTags) > 0 {
		filters.Tags = input.FilterTags
	}

	entries, err := uc.entryRepo.GetAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	categoryMap := make(map[string]*entity.Category)

	accounts, err := uc.accountRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	for _, acc := range accounts {
		categories, err := uc.categoryRepo.GetAll(acc.Name)
		if err != nil {
			continue
		}
		for _, cat := range categories {
			categoryMap[cat.Alias] = cat
		}
	}

	output := &ReportOutput{
		Entries: make([]*EntryWithCategory, 0),
	}

	for _, entry := range entries {
		if len(input.FilterCategories) > 0 {
			if entry.CategoryAlias == nil {
				continue
			}
			cat, ok := categoryMap[*entry.CategoryAlias]
			if !ok {
				continue
			}
			if !uc.matchesCategoryFilter(cat, input.FilterCategories) {
				continue
			}
		}

		entryWithCat := &EntryWithCategory{
			Type:              string(entry.Type),
			Amount:            entry.Amount,
			Currency:          entry.Currency,
			Description:       entry.Description,
			CategoryAlias:     entry.CategoryAlias,
			CreditCardName:    entry.CreditCardName,
			AccountName:       input.AccountName,
			InstallmentNumber: entry.InstallmentNumber,
			InstallmentTotal:  entry.InstallmentTotal,
			RealizationDate:   entry.RealizationDate,
			PaymentDate:       entry.PaymentDate,
			Tags:              entry.Tags,
		}

		if entry.CategoryAlias != nil {
			if cat, ok := categoryMap[*entry.CategoryAlias]; ok {
				entryWithCat.CategoryName = cat.Name
			}
		}

		if entry.IsIncome() {
			output.TotalIncome += entry.Amount
		} else {
			output.TotalExpense += entry.Amount
		}

		output.Entries = append(output.Entries, entryWithCat)
	}

	output.Accounts = accounts

	return output, nil
}

func (uc *Report) matchesCategoryFilter(cat *entity.Category, filter []string) bool {
	for _, f := range filter {
		if cat.Name == f {
			return true
		}
		if cat.Alias == f {
			return true
		}
	}
	return false
}
