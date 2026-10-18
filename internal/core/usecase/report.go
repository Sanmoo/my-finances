package usecase

import (
	"fmt"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type ReportInput struct {
	From                    *time.Time
	To                      *time.Time
	FilterTags              []string
	FilterCategories        []string
	AccountName             string
	FilterByRealizationDate bool
}

// ReportOutput carries everything the callers need to render the report:
// the entries in repository order, the legacy category map keyed by alias
// (see Execute for its exact resolution semantics) and the accounts by name.
type ReportOutput struct {
	Entries        []*entity.Entry
	Categories     map[string]*entity.Category
	AccountsByName map[string]*entity.Account
	TotalIncome    float64
	TotalExpense   float64
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

// Execute queries the entries, resolves Categories for every account and
// computes the income/expense totals.
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

	accounts, err := uc.accountRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	categoryMap := buildLegacyCategoryMap(uc.categoryRepo, accounts)

	accountsByName := make(map[string]*entity.Account, len(accounts))
	for _, acc := range accounts {
		accountsByName[acc.Name] = acc
	}

	output := &ReportOutput{
		Entries:        make([]*entity.Entry, 0, len(entries)),
		Categories:     categoryMap,
		AccountsByName: accountsByName,
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
			if !matchesCategoryFilter(cat, input.FilterCategories) {
				continue
			}
		}

		if entry.IsIncome() {
			output.TotalIncome += entry.Amount
		} else {
			output.TotalExpense += entry.Amount
		}

		output.Entries = append(output.Entries, entry)
	}

	return output, nil
}

// buildLegacyCategoryMap resolves Categories per account into the legacy map
// keyed by alias: accounts are iterated in repository order (even when a
// specific account filters the entries), the LAST account loaded wins on
// alias collisions, read errors are ignored and a missing file is just an
// absent set. This is the deliberately preserved legacy behavior of issue #3.
func buildLegacyCategoryMap(categoryRepo port.CategoriesRepository, accounts []*entity.Account) map[string]*entity.Category {
	categoryMap := make(map[string]*entity.Category)

	for _, acc := range accounts {
		categories, err := categoryRepo.GetAll(acc.Name)
		if err != nil {
			continue
		}
		for _, cat := range categories {
			categoryMap[cat.Alias] = cat
		}
	}

	return categoryMap
}

func matchesCategoryFilter(cat *entity.Category, filter []string) bool {
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
