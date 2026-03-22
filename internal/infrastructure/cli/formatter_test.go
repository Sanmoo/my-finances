package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/i18n"
	"github.com/stretchr/testify/assert"
)

func intPtr(i int64) *int64 {
	return &i
}

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
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: longName, Type: entity.CategoryTypeExpense},
		}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)

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
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          50.00,
				Currency:        "BRL",
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: "food", Emoji: &emoji, Type: entity.CategoryTypeExpense},
		}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)

		assert.Contains(t, output, emoji)
		assert.Contains(t, output, "food")
	})

	t.Run("uses minimum width Category when no categories", func(t *testing.T) {
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)

		assert.Contains(t, output, "Category")
	})

	t.Run("adjusts width for very long category names", func(t *testing.T) {
		veryLongName := "assinaturas de serviços e aplicativos diversos"
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: veryLongName, Type: entity.CategoryTypeExpense},
		}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)

		assert.Contains(t, output, veryLongName)
	})

	t.Run("calculates width correctly with multiple categories", func(t *testing.T) {
		shortName := "food"
		longName := "transporte & derivados"
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          50.00,
				Currency:        "BRL",
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
			{
				ID:              2,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				CategoryID:      intPtr(2),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: shortName, Type: entity.CategoryTypeExpense},
			2: {ID: 2, Name: longName, Type: entity.CategoryTypeExpense},
		}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)

		assert.Contains(t, output, shortName)
		assert.Contains(t, output, longName)
	})
}

func TestFormatEntriesTable_SeparatorLength(t *testing.T) {
	f := newTestFormatter()

	t.Run("separator matches header width for long categories", func(t *testing.T) {
		longName := "transporte & derivados muito longo"
		entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				Currency:        "BRL",
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: longName, Type: entity.CategoryTypeExpense},
		}

		output := f.FormatEntriesTable(entries, categories, nil, nil, nil, nil)
		lines := strings.Split(output, "\n")

		var headerLine, separatorLine string
		for i, line := range lines {
			if strings.Contains(line, "Date") && strings.Contains(line, "Category") {
				headerLine = line
				if i+1 < len(lines) {
					separatorLine = lines[i+1]
				}
				break
			}
		}

		assert.NotEmpty(t, headerLine, "should find header line")
		assert.NotEmpty(t, separatorLine, "should find separator line")
		assert.Len(t, separatorLine, len(headerLine), "separator should match header length")
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
			ID:   1,
			Name: "food",
			Type: entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "food", result)
	})

	t.Run("returns name without emoji when emoji is empty", func(t *testing.T) {
		emptyEmoji := ""
		cat := &entity.Category{
			ID:    1,
			Name:  "food",
			Emoji: &emptyEmoji,
			Type:  entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "food", result)
	})

	t.Run("returns emoji and name when emoji is present", func(t *testing.T) {
		emoji := "🍕"
		cat := &entity.Category{
			ID:    1,
			Name:  "food",
			Emoji: &emoji,
			Type:  entity.CategoryTypeExpense,
		}
		result := f.getCategoryDisplayName(cat)
		assert.Equal(t, "🍕 food", result)
	})
}

func TestCalculateCategoryWidth(t *testing.T) {
	f := newTestFormatter()
	entryDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	t.Run("returns minimum width when no entries", func(t *testing.T) {
		entries := []*entity.Entry{}
		categories := map[int64]*entity.Category{}

		width := f.calculateCategoryWidth(entries, categories)

		assert.Equal(t, len("Category"), width)
	})

	t.Run("returns minimum width when entries have no category", func(t *testing.T) {
		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{}

		width := f.calculateCategoryWidth(entries, categories)

		assert.Equal(t, len("Category"), width)
	})

	t.Run("returns category name length for longer names", func(t *testing.T) {
		longName := "transporte & derivados"
		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: longName, Type: entity.CategoryTypeExpense},
		}

		width := f.calculateCategoryWidth(entries, categories)

		assert.Equal(t, len(longName), width)
	})

	t.Run("includes emoji in width calculation", func(t *testing.T) {
		emoji := "🍕"
		name := "food"
		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: name, Emoji: &emoji, Type: entity.CategoryTypeExpense},
		}

		width := f.calculateCategoryWidth(entries, categories)

		expectedLen := len(emoji) + len(" ") + len(name)
		assert.Equal(t, expectedLen, width)
	})

	t.Run("returns longest category width from multiple entries", func(t *testing.T) {
		shortName := "food"
		longName := "transporte & derivados"
		entries := []*entity.Entry{
			{
				ID:              1,
				Type:            entity.EntryTypeExpense,
				Amount:          50.00,
				CategoryID:      intPtr(1),
				RealizationDate: entryDate,
			},
			{
				ID:              2,
				Type:            entity.EntryTypeExpense,
				Amount:          100.00,
				CategoryID:      intPtr(2),
				RealizationDate: entryDate,
			},
		}
		categories := map[int64]*entity.Category{
			1: {ID: 1, Name: shortName, Type: entity.CategoryTypeExpense},
			2: {ID: 2, Name: longName, Type: entity.CategoryTypeExpense},
		}

		width := f.calculateCategoryWidth(entries, categories)

		assert.Equal(t, len(longName), width)
	})
}
