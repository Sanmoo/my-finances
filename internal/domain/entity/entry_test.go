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
			name:      "valid negative expense entry",
			entryType: EntryTypeExpense,
			amount:    -20.00,
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
			name:      "negative income amount",
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

	t.Run("with category alias", func(t *testing.T) {
		alias := "food"
		entry, err := NewEntry(EntryTypeIncome, 1000.00, "BRL", baseDate, WithCategoryAlias(alias))
		assert.NoError(t, err)
		assert.NotNil(t, entry.CategoryAlias)
		assert.Equal(t, alias, *entry.CategoryAlias)
	})

	t.Run("with credit card calculates payment date", func(t *testing.T) {
		cc := &CreditCard{
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

	t.Run("with tags", func(t *testing.T) {
		tags := []string{"tag1", "tag2"}
		entry, err := NewEntry(EntryTypeExpense, 50.00, "BRL", baseDate, WithTags(tags))
		assert.NoError(t, err)
		assert.Equal(t, tags, entry.Tags)
	})

	t.Run("with installment", func(t *testing.T) {
		entry, err := NewEntry(EntryTypeExpense, 50.00, "BRL", baseDate, WithInstallment(1, 5))
		assert.NoError(t, err)
		assert.Equal(t, 1, entry.InstallmentNumber)
		assert.Equal(t, 5, entry.InstallmentTotal)
	})
}

func TestEntry_AddTag(t *testing.T) {
	entry := &Entry{Tags: []string{"tag1"}}

	entry.AddTag("tag2")
	assert.Equal(t, []string{"tag1", "tag2"}, entry.Tags)

	entry.AddTag("tag1")
	assert.Equal(t, []string{"tag1", "tag2"}, entry.Tags)
}

func TestEntry_IsExpense(t *testing.T) {
	entry := &Entry{Type: EntryTypeExpense}
	assert.True(t, entry.IsExpense())
	assert.False(t, entry.IsIncome())

	entry.Type = EntryTypeIncome
	assert.False(t, entry.IsExpense())
	assert.True(t, entry.IsIncome())
}

func TestNewEntry_WithCreditCardInstallments(t *testing.T) {
	cc := &CreditCard{
		Name:       "Test Card",
		ClosingDay: 9,
		DueDay:     16,
	}

	realizationDate := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)

	t.Run("installment 1 of 6 has correct payment date", func(t *testing.T) {
		entry, err := NewEntry(
			EntryTypeExpense,
			66.42,
			"BRL",
			realizationDate,
			WithInstallment(1, 6),
			WithCreditCard(cc),
		)
		assert.NoError(t, err)
		assert.Equal(t, realizationDate, entry.RealizationDate)
		assert.NotNil(t, entry.PaymentDate)
		assert.Equal(t, 16, entry.PaymentDate.Day())
		assert.Equal(t, time.Month(4), entry.PaymentDate.Month())
		assert.Equal(t, 2026, entry.PaymentDate.Year())
	})

	t.Run("installment 2 of 6 has correct payment date", func(t *testing.T) {
		entry, err := NewEntry(
			EntryTypeExpense,
			66.42,
			"BRL",
			realizationDate,
			WithInstallment(2, 6),
			WithCreditCard(cc),
		)
		assert.NoError(t, err)
		assert.Equal(t, realizationDate, entry.RealizationDate)
		assert.NotNil(t, entry.PaymentDate)
		assert.Equal(t, 16, entry.PaymentDate.Day())
		assert.Equal(t, time.Month(5), entry.PaymentDate.Month())
		assert.Equal(t, 2026, entry.PaymentDate.Year())
	})

	t.Run("installment 6 of 6 has correct payment date", func(t *testing.T) {
		entry, err := NewEntry(
			EntryTypeExpense,
			66.42,
			"BRL",
			realizationDate,
			WithInstallment(6, 6),
			WithCreditCard(cc),
		)
		assert.NoError(t, err)
		assert.Equal(t, realizationDate, entry.RealizationDate)
		assert.NotNil(t, entry.PaymentDate)
		assert.Equal(t, 16, entry.PaymentDate.Day())
		assert.Equal(t, time.Month(9), entry.PaymentDate.Month())
		assert.Equal(t, 2026, entry.PaymentDate.Year())
	})

	t.Run("all installments have same realization date", func(t *testing.T) {
		for i := 1; i <= 6; i++ {
			entry, err := NewEntry(
				EntryTypeExpense,
				66.42,
				"BRL",
				realizationDate,
				WithInstallment(i, 6),
				WithCreditCard(cc),
			)
			assert.NoError(t, err)
			assert.Equal(t, realizationDate, entry.RealizationDate,
				"installment %d should have same realization date", i)
		}
	})

	t.Run("order of options matters - WithInstallment must come before WithCreditCard", func(t *testing.T) {
		entryWrong, err := NewEntry(
			EntryTypeExpense,
			66.42,
			"BRL",
			realizationDate,
			WithCreditCard(cc),
			WithInstallment(2, 6),
		)
		assert.NoError(t, err)
		assert.NotNil(t, entryWrong.PaymentDate)
	})
}
