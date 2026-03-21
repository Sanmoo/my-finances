package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEntry(t *testing.T) {
	baseDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		entryType EntryType
		amount    float64
		currency  string
		date      time.Time
		opts      []EntryOption
		wantErr   error
	}{
		{
			name:      "valid income entry",
			entryType: EntryTypeIncome,
			amount:    1000.50,
			currency:  "BRL",
			date:      baseDate,
			opts:      []EntryOption{},
			wantErr:   nil,
		},
		{
			name:      "valid expense entry",
			entryType: EntryTypeExpense,
			amount:    50.00,
			currency:  "BRL",
			date:      baseDate,
			opts:      []EntryOption{},
			wantErr:   nil,
		},
		{
			name:      "invalid entry type",
			entryType: "invalid",
			amount:    100.00,
			currency:  "BRL",
			date:      baseDate,
			wantErr:   ErrInvalidEntryType,
		},
		{
			name:      "zero amount",
			entryType: EntryTypeIncome,
			amount:    0,
			currency:  "BRL",
			date:      baseDate,
			wantErr:   ErrInvalidAmount,
		},
		{
			name:      "negative amount",
			entryType: EntryTypeIncome,
			amount:    -10.00,
			currency:  "BRL",
			date:      baseDate,
			wantErr:   ErrInvalidAmount,
		},
		{
			name:      "empty currency",
			entryType: EntryTypeIncome,
			amount:    100.00,
			currency:  "",
			date:      baseDate,
			wantErr:   ErrEmptyCurrency,
		},
		{
			name:      "zero date",
			entryType: EntryTypeIncome,
			amount:    100.00,
			currency:  "BRL",
			date:      time.Time{},
			wantErr:   ErrInvalidDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := NewEntry(tt.entryType, tt.amount, tt.currency, tt.date, tt.opts...)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
				assert.Equal(t, tt.entryType, entry.Type)
				assert.Equal(t, tt.amount, entry.Amount)
				assert.Equal(t, tt.currency, entry.Currency)
				assert.False(t, entry.CreatedAt.IsZero())
			}
		})
	}
}

func TestNewEntry_WithOptions(t *testing.T) {
	baseDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	t.Run("with description", func(t *testing.T) {
		entry, err := NewEntry(EntryTypeExpense, 50.00, "BRL", baseDate, WithDescription("Lunch"))
		assert.NoError(t, err)
		assert.Equal(t, "Lunch", entry.Description)
	})

	t.Run("with category ID", func(t *testing.T) {
		catID := int64(5)
		entry, err := NewEntry(EntryTypeIncome, 1000.00, "BRL", baseDate, WithCategoryID(catID))
		assert.NoError(t, err)
		assert.NotNil(t, entry.CategoryID)
		assert.Equal(t, catID, *entry.CategoryID)
	})

	t.Run("with credit card calculates payment date", func(t *testing.T) {
		cc := &CreditCard{
			ID:         1,
			Name:       "Test Card",
			ClosingDay: 9,
			DueDay:     16,
		}

		dateBeforeClosing := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
		entry, err := NewEntry(EntryTypeExpense, 100.00, "BRL", dateBeforeClosing, WithCreditCard(cc))
		assert.NoError(t, err)
		assert.NotNil(t, entry.PaymentDate)
		assert.Equal(t, 16, entry.PaymentDate.Day())
		assert.Equal(t, time.Month(3), entry.PaymentDate.Month())
	})

	t.Run("with tag IDs", func(t *testing.T) {
		tagIDs := []int64{1, 2}
		entry, err := NewEntry(EntryTypeExpense, 50.00, "BRL", baseDate, WithTagIDs(tagIDs))
		assert.NoError(t, err)
		assert.Equal(t, tagIDs, entry.TagIDs)
	})
}

func TestEntry_AddTagID(t *testing.T) {
	entry := &Entry{TagIDs: []int64{1}}

	entry.AddTagID(2)
	assert.Equal(t, []int64{1, 2}, entry.TagIDs)

	entry.AddTagID(1)
	assert.Equal(t, []int64{1, 2}, entry.TagIDs)
}

func TestEntry_IsExpense(t *testing.T) {
	entry := &Entry{Type: EntryTypeExpense}
	assert.True(t, entry.IsExpense())
	assert.False(t, entry.IsIncome())

	entry.Type = EntryTypeIncome
	assert.False(t, entry.IsExpense())
	assert.True(t, entry.IsIncome())
}
