package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCreditCard(t *testing.T) {
	tests := []struct {
		name       string
		ccName     string
		closingDay int
		dueDay     int
		wantErr    error
	}{
		{
			name:       "valid credit card",
			ccName:     "Nubank",
			closingDay: 9,
			dueDay:     16,
			wantErr:    nil,
		},
		{
			name:       "closing day too low",
			ccName:     "Nubank",
			closingDay: 0,
			dueDay:     16,
			wantErr:    ErrInvalidClosingDay,
		},
		{
			name:       "closing day too high",
			ccName:     "Nubank",
			closingDay: 32,
			dueDay:     16,
			wantErr:    ErrInvalidClosingDay,
		},
		{
			name:       "due day too low",
			ccName:     "Nubank",
			closingDay: 9,
			dueDay:     0,
			wantErr:    ErrInvalidDueDay,
		},
		{
			name:       "due day too high",
			ccName:     "Nubank",
			closingDay: 9,
			dueDay:     32,
			wantErr:    ErrInvalidDueDay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, err := NewCreditCard(tt.ccName, tt.closingDay, tt.dueDay)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cc)
				assert.Equal(t, tt.ccName, cc.Name)
				assert.Equal(t, tt.closingDay, cc.ClosingDay)
				assert.Equal(t, tt.dueDay, cc.DueDay)
			}
		})
	}
}

func TestCreditCard_CalculatePaymentDate(t *testing.T) {
	cc := &CreditCard{
		ID:         1,
		Name:       "Test Card",
		ClosingDay: 9,
		DueDay:     16,
	}

	tests := []struct {
		name            string
		realizationDate time.Time
		expectedMonth   time.Month
		expectedDay     int
	}{
		{
			name:            "day before closing day",
			realizationDate: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			expectedMonth:   3,
			expectedDay:     16,
		},
		{
			name:            "day at closing day",
			realizationDate: time.Date(2024, 3, 9, 0, 0, 0, 0, time.UTC),
			expectedMonth:   3,
			expectedDay:     16,
		},
		{
			name:            "day after closing day",
			realizationDate: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
			expectedMonth:   4,
			expectedDay:     16,
		},
		{
			name:            "last day of month before closing",
			realizationDate: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			expectedMonth:   2,
			expectedDay:     16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentDate := cc.CalculatePaymentDate(tt.realizationDate)
			assert.Equal(t, tt.expectedMonth, paymentDate.Month())
			assert.Equal(t, tt.expectedDay, paymentDate.Day())
		})
	}
}

func TestCreditCard_CalculateInstallmentPaymentDate(t *testing.T) {
	cc := &CreditCard{
		ID:         1,
		Name:       "Test Card",
		ClosingDay: 9,
		DueDay:     16,
	}

	realizationDate := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)

	installment1 := cc.CalculateInstallmentPaymentDate(realizationDate, 1)
	assert.Equal(t, time.Month(4), installment1.Month())
	assert.Equal(t, 16, installment1.Day())

	installment2 := cc.CalculateInstallmentPaymentDate(realizationDate, 2)
	assert.Equal(t, time.Month(5), installment2.Month())
	assert.Equal(t, 16, installment2.Day())

	installment3 := cc.CalculateInstallmentPaymentDate(realizationDate, 3)
	assert.Equal(t, time.Month(6), installment3.Month())
	assert.Equal(t, 16, installment3.Day())
}
