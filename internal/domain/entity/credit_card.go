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
