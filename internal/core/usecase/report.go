package usecase

import (
	"fmt"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type ReportInput struct {
	NamespaceID      int64
	Format           string
	From             *time.Time
	To               *time.Time
	FilterTags       []string
	FilterCategories []string
}

type ReportOutput struct {
	Entries      []*EntryWithCategory
	TotalIncome  float64
	TotalExpense float64
}

type EntryWithCategory struct {
	EntryID         int64
	Type            string
	Amount          float64
	Currency        string
	Description     string
	CategoryID      *int64
	CategoryName    string
	RealizationDate time.Time
	PaymentDate     *time.Time
	Tags            []string
}

type Report struct {
	entryRepo    port.EntriesRepository
	categoryRepo port.CategoriesRepository
}

func NewReport(entryRepo port.EntriesRepository, categoryRepo port.CategoriesRepository) *Report {
	return &Report{
		entryRepo:    entryRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *Report) Execute(input ReportInput) (*ReportOutput, error) {
	filters := &port.EntryFilters{
		FromDate: input.From,
		ToDate:   input.To,
	}

	if len(input.FilterTags) > 0 {
		filters.Tags = input.FilterTags
	}

	entries, err := uc.entryRepo.GetByNamespaceID(input.NamespaceID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	categories, err := uc.categoryRepo.GetByNamespaceID(input.NamespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	categoryMap := make(map[int64]*entity.Category)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}

	output := &ReportOutput{
		Entries: make([]*EntryWithCategory, 0),
	}

	for _, entry := range entries {
		if len(input.FilterCategories) > 0 {
			if entry.CategoryID == nil {
				continue
			}
			cat, ok := categoryMap[*entry.CategoryID]
			if !ok {
				continue
			}
			if !uc.matchesCategoryFilter(cat, input.FilterCategories) {
				continue
			}
		}

		entryWithCat := &EntryWithCategory{
			EntryID:         entry.ID,
			Type:            string(entry.Type),
			Amount:          entry.Amount,
			Currency:        entry.Currency,
			Description:     entry.Description,
			CategoryID:      entry.CategoryID,
			RealizationDate: entry.RealizationDate,
			PaymentDate:     entry.PaymentDate,
			Tags:            entry.Tags,
		}

		if entry.CategoryID != nil {
			if cat, ok := categoryMap[*entry.CategoryID]; ok {
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

	return output, nil
}

func (uc *Report) matchesCategoryFilter(cat *entity.Category, filter []string) bool {
	for _, f := range filter {
		if cat.Name == f {
			return true
		}
		if cat.Alias != nil && *cat.Alias == f {
			return true
		}
	}
	return false
}
