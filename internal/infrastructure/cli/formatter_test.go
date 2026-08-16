package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/usecase"
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

func statementTestOutput() *usecase.StatementOutput {
	entry1 := &entity.Entry{
		Type:              entity.EntryTypeExpense,
		Amount:            100.00,
		Currency:          "BRL",
		Description:       "May purchase",
		CategoryAlias:     strPtr("rest"),
		CreditCardName:    strPtr("main"),
		Tags:              []string{"credito"},
		InstallmentNumber: 1,
		InstallmentTotal:  2,
		RealizationDate:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
	}
	entry2 := &entity.Entry{
		Type:            entity.EntryTypeExpense,
		Amount:          50.50,
		Currency:        "BRL",
		Description:     "June purchase",
		CategoryAlias:   strPtr("rest"),
		CreditCardName:  strPtr("main"),
		RealizationDate: time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
	}
	return &usecase.StatementOutput{
		Card:        &entity.CreditCard{Name: "main", ClosingDay: 9, DueDay: 16},
		Month:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		WindowStart: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC),
		WindowEnd:   time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC),
		DueDate:     time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		Lines: []usecase.StatementLine{
			{Entry: entry1, AccountName: "main"},
			{Entry: entry2, AccountName: "deh"},
		},
		Totals: []usecase.CurrencyTotal{{Currency: "BRL", Amount: 150.50}},
	}
}

func TestFormatStatementTable(t *testing.T) {
	f := newTestFormatter()
	categories := map[string]*entity.Category{
		"rest": {Name: "Restaurante", Alias: "rest", Emoji: strPtr("🍽"), Type: entity.CategoryTypeExpense},
	}

	t.Run("prints header with card, month, window, due date and total", func(t *testing.T) {
		output := f.FormatStatementTable(statementTestOutput(), categories)

		assert.Contains(t, output, `Invoice for card "main" — June 2026`)
		assert.Contains(t, output, "Purchase window: 09/05/2026 to 08/06/2026")
		assert.Contains(t, output, "Due date: 16/06/2026")
		assert.Contains(t, output, "Total: R$ 150,50")
	})

	t.Run("prints rows with realization date, category, description, installment and tags", func(t *testing.T) {
		output := f.FormatStatementTable(statementTestOutput(), categories)

		assert.Contains(t, output, "20/05/2026")
		assert.Contains(t, output, "🍽 Restaurante")
		assert.Contains(t, output, "May purchase (1/2) [credito]")
	})

	t.Run("shows account column when lines span multiple accounts", func(t *testing.T) {
		output := f.FormatStatementTable(statementTestOutput(), categories)

		assert.Contains(t, output, "Account")
		assert.Contains(t, output, "main")
		assert.Contains(t, output, "deh")
	})

	t.Run("omits account column for a single account", func(t *testing.T) {
		out := statementTestOutput()
		out.Lines = out.Lines[:1]

		output := f.FormatStatementTable(out, categories)

		assert.NotContains(t, output, "Account")
		assert.NotContains(t, output, "deh")
	})

	t.Run("prints empty message when there are no lines", func(t *testing.T) {
		out := statementTestOutput()
		out.Lines = nil
		out.Totals = nil

		output := f.FormatStatementTable(out, categories)

		assert.Contains(t, output, "No entries for this invoice")
	})

	t.Run("prints one total per currency", func(t *testing.T) {
		out := statementTestOutput()
		out.Totals = []usecase.CurrencyTotal{
			{Currency: "BRL", Amount: 150.50},
			{Currency: "USD", Amount: 20.00},
		}

		output := f.FormatStatementTable(out, categories)

		assert.Contains(t, output, "Total: R$ 150,50 | $ 20,00")
	})
}

func TestFormatStatementMarkdown(t *testing.T) {
	f := newTestFormatter()
	categories := map[string]*entity.Category{
		"rest": {Name: "Restaurante", Alias: "rest", Emoji: strPtr("🍽"), Type: entity.CategoryTypeExpense},
	}

	t.Run("prints markdown header, rows and total", func(t *testing.T) {
		output := f.FormatStatementMarkdown(statementTestOutput(), categories)

		assert.Contains(t, output, `# Invoice for card "main" — June 2026`)
		assert.Contains(t, output, "**Purchase window:** 09/05/2026 to 08/06/2026")
		assert.Contains(t, output, "**Due date:** 16/06/2026")
		assert.Contains(t, output, "**Total:** R$ 150,50")
		assert.Contains(t, output, "| 20/05/2026 |")
		assert.Contains(t, output, "May purchase (1/2) [credito]")
	})

	t.Run("prints empty message when there are no lines", func(t *testing.T) {
		out := statementTestOutput()
		out.Lines = nil
		out.Totals = nil

		output := f.FormatStatementMarkdown(out, categories)

		assert.Contains(t, output, "No entries for this invoice")
	})
}
