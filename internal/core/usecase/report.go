package usecase

import (
	"fmt"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type ReportInput struct {
	Format           string
	From             *time.Time
	To               *time.Time
	FilterTags       []string
	FilterCategories []string
	AccountID        *int64
}

type ReportOutput struct {
	Entries           []*EntryWithCategory
	TotalIncome       float64
	TotalExpense      float64
	TotalInstallments map[int64]int
}

type EntryWithCategory struct {
	EntryID         int64
	Type            string
	Amount          float64
	Currency        string
	Description     string
	CategoryID      *int64
	CategoryName    string
	CreditCardID    *int64
	AccountID       int64
	Installment     int
	ParentEntryID   *int64
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
		FromDate:  input.From,
		ToDate:    input.To,
		AccountID: input.AccountID,
	}

	if len(input.FilterTags) > 0 {
		filters.Tags = input.FilterTags
	}

	entries, err := uc.entryRepo.GetAll(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	categoryMap := make(map[int64]*entity.Category)

	if input.AccountID != nil {
		categories, err := uc.categoryRepo.GetAll(*input.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get categories: %w", err)
		}
		for _, cat := range categories {
			categoryMap[cat.ID] = cat
		}
	}

	totalInstallments := uc.calculateTotalInstallments(entries)

	output := &ReportOutput{
		Entries:           make([]*EntryWithCategory, 0),
		TotalInstallments: totalInstallments,
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
			CreditCardID:    entry.CreditCardID,
			AccountID:       entry.AccountID,
			Installment:     entry.Installment,
			ParentEntryID:   entry.ParentEntryID,
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

func (uc *Report) calculateTotalInstallments(entries []*entity.Entry) map[int64]int {
	result := make(map[int64]int)

	for _, entry := range entries {
		var groupID int64
		if entry.ParentEntryID != nil {
			groupID = *entry.ParentEntryID
		} else {
			groupID = entry.ID
		}
		result[groupID]++
	}

	return result
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
