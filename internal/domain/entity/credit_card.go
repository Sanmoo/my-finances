package entity

import (
	"errors"
	"time"
)

var (
	ErrInvalidClosingDay = errors.New("closing_day must be between 1 and 31")
	ErrInvalidDueDay     = errors.New("due_day must be between 1 and 31")
)

type CreditCard struct {
	Name       string
	ClosingDay int
	DueDay     int
}

func NewCreditCard(name string, closingDay, dueDay int) (*CreditCard, error) {
	if closingDay < 1 || closingDay > 31 {
		return nil, ErrInvalidClosingDay
	}
	if dueDay < 1 || dueDay > 31 {
		return nil, ErrInvalidDueDay
	}

	return &CreditCard{
		Name:       name,
		ClosingDay: closingDay,
		DueDay:     dueDay,
	}, nil
}

func (cc *CreditCard) CalculatePaymentDate(realizationDate time.Time) time.Time {
	year := realizationDate.Year()
	month := int(realizationDate.Month())
	day := realizationDate.Day()

	if day < cc.ClosingDay {
		return time.Date(year, time.Month(month), cc.DueDay, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(year, time.Month(month+1), cc.DueDay, 0, 0, 0, 0, time.UTC)
}

func (cc *CreditCard) CalculateInstallmentPaymentDate(realizationDate time.Time, installment int) time.Time {
	baseDate := cc.CalculatePaymentDate(realizationDate)
	return baseDate.AddDate(0, installment-1, 0)
}

// InvoiceWindow returns the realization-date interval covered by the invoice
// due in the given month: purchases from the previous month's closing day
// through the day before the current month's closing day. When the closing
// day does not exist in a month (e.g. day 31 in February), the window is
// clamped to that month's last day, consistent with CalculatePaymentDate.
func (cc *CreditCard) InvoiceWindow(year int, month time.Month) (start, end time.Time) {
	prevMonthLastDay := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
	monthLastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)

	if cc.ClosingDay > prevMonthLastDay.Day() {
		start = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	} else {
		start = time.Date(year, month-1, cc.ClosingDay, 0, 0, 0, 0, time.UTC)
	}

	switch {
	case cc.ClosingDay <= 1:
		end = prevMonthLastDay
	case cc.ClosingDay-1 <= monthLastDay.Day():
		end = time.Date(year, month, cc.ClosingDay-1, 0, 0, 0, 0, time.UTC)
	default:
		end = monthLastDay
	}

	return start, end
}

// DueDate returns the due date of the invoice for the given month, using the
// same day normalization as CalculatePaymentDate.
func (cc *CreditCard) DueDate(year int, month time.Month) time.Time {
	return time.Date(year, month, cc.DueDay, 0, 0, 0, 0, time.UTC)
}
