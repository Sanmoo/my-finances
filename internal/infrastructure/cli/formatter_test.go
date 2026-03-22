package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/i18n"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string {
	return &s
}

func newTestFormatter() *Formatter {
	return NewFormatter(i18n.New("pt-BR"))
}

func TestFormatEntriesTable_CategoryWidth(t *testing.T) {
	f := newTestFormatter()

	t.Run("adjusts width for long category names", func(t *testing.T) {
		longName := "transporte & derivados muito longo"
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				CategoryAlias:   strPtr("transport"),
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{
			"transport": {Name: longName, Alias: "transport", Type: entity.CategoryTypeExpense},
		}
		accounts := map[string]*entity.Account{
			"test": {Name: "test"},
		}

		output := f.FormatEntriesTable(entries, categories, accounts, "test")

		assert.Contains(t, output, longName)
		lines := strings.Split(output, "\n")
		var dataLine string
		for _, line := range lines {
			if strings.Contains(line, longName) {
				dataLine = line
				break
			}
		}
		assert.NotEmpty(t, dataLine, "should find line with category name")
	})

	t.Run("handles emoji prefix in width calculation", func(t *testing.T) {
		emoji := "🍕"
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          50.00,
				Currency:        "BRL",
				CategoryAlias:   strPtr("food"),
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{
			"food": {Name: "food", Alias: "food", Emoji: &emoji, Type: entity.CategoryTypeExpense},
		}
		accounts := map[string]*entity.Account{
			"test": {Name: "test"},
		}

		output := f.FormatEntriesTable(entries, categories, accounts, "test")

		assert.Contains(t, output, emoji)
		assert.Contains(t, output, "food")
	})

	t.Run("uses minimum width Category when no categories", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{
			"test": {Name: "test"},
		}

		output := f.FormatEntriesTable(entries, categories, accounts, "test")

		assert.Contains(t, output, "15/03/2024")
		assert.Contains(t, output, "R$ 100,00")
	})

	t.Run("prints headers for expenses", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesTable(entries, categories, accounts, "")

		assert.Contains(t, output, "=== Expenses ===")
		assert.Contains(t, output, "Date")
		assert.Contains(t, output, "Category")
		assert.Contains(t, output, "Amount")
		assert.Contains(t, output, "Description")
	})

	t.Run("prints headers for incomes", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeIncome,
				Amount:          1000.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesTable(entries, categories, accounts, "")

		assert.Contains(t, output, "=== Incomes ===")
		assert.Contains(t, output, "Date")
		assert.Contains(t, output, "Category")
		assert.Contains(t, output, "Amount")
		assert.Contains(t, output, "Description")
	})

	t.Run("separates expenses and incomes", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
				Description:     "Test expense",
			},
			{
				Type:            entity.EntryTypeIncome,
				Amount:          1000.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
				Description:     "Test income",
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesTable(entries, categories, accounts, "")

		expensesIdx := strings.Index(output, "=== Expenses ===")
		incomesIdx := strings.Index(output, "=== Incomes ===")

		assert.True(t, expensesIdx < incomesIdx, "Expenses should come before Incomes")
		assert.Contains(t, output, "Test expense")
		assert.Contains(t, output, "Test income")
	})
}

func TestFormatEntriesMarkdown_HeadersAndSeparation(t *testing.T) {
	f := newTestFormatter()

	t.Run("prints markdown headers for expenses", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesMarkdown(entries, categories, accounts, "")

		assert.Contains(t, output, "## Expenses")
		assert.Contains(t, output, "| Date | Category | Amount | Description |")
	})

	t.Run("prints markdown headers for incomes", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeIncome,
				Amount:          1000.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesMarkdown(entries, categories, accounts, "")

		assert.Contains(t, output, "## Incomes")
		assert.Contains(t, output, "| Date | Category | Amount | Description |")
	})

	t.Run("separates expenses and incomes in markdown", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
				Description:     "Test expense",
			},
			{
				Type:            entity.EntryTypeIncome,
				Amount:          1000.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
				Description:     "Test income",
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesMarkdown(entries, categories, accounts, "")

		expensesIdx := strings.Index(output, "## Expenses")
		incomesIdx := strings.Index(output, "## Incomes")

		assert.True(t, expensesIdx < incomesIdx, "Expenses should come before Incomes")
		assert.Contains(t, output, "Test expense")
		assert.Contains(t, output, "Test income")
	})
}

func TestGetCategoryDisplayName(t *testing.T) {
	f := newTestFormatter()

	t.Run("returns empty string for nil category", func(t *testing.T) {
		result := f.getCategoryDisplayName(nil)
		assert.Empty(t, result)
	})

	t.Run("returns name without emoji when emoji is nil", func(t *testing.T) {
		cat := &entity.Category{
			Name:  "food",
			Alias: "food",
			Type:  entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "food", result)
	})

	t.Run("returns name without emoji when emoji is empty", func(t *testing.T) {
		emptyEmoji := ""
		cat := &entity.Category{
			Name:  "food",
			Alias: "food",
			Emoji: &emptyEmoji,
			Type:  entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "food", result)
	})

	t.Run("returns emoji and name when emoji is present", func(t *testing.T) {
		emoji := "🍕"
		cat := &entity.Category{
			Name:  "food",
			Alias: "food",
			Emoji: &emoji,
			Type:  entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "🍕 food", result)
	})
}

func TestFormatEntriesTable_PaymentDate(t *testing.T) {
	f := newTestFormatter()

	t.Run("shows payment date when available", func(t *testing.T) {
		realizationDate := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
		paymentDate := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: realizationDate,
				PaymentDate:     &paymentDate,
				Description:     "Test credit card purchase",
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesTable(entries, categories, accounts, "")

		assert.Contains(t, output, "16/04/2026")
	})

	t.Run("shows realization date when payment date is nil", func(t *testing.T) {
		realizationDate := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: realizationDate,
				Description:     "Test normal purchase",
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesTable(entries, categories, accounts, "")

		assert.Contains(t, output, "14/03/2026")
	})
}

func TestFormatEntriesMarkdown_PaymentDate(t *testing.T) {
	f := newTestFormatter()

	t.Run("shows payment date in markdown when available", func(t *testing.T) {
		realizationDate := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
		paymentDate := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: realizationDate,
				PaymentDate:     &paymentDate,
				Description:     "Test credit card purchase",
			},
		}
		categories := map[string]*entity.Category{}
		accounts := map[string]*entity.Account{}

		output := f.FormatEntriesMarkdown(entries, categories, accounts, "")

		assert.Contains(t, output, "16/04/2026")
	})
}
