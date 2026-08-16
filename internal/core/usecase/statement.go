package usecase

import (
	"fmt"
	"sort"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type StatementInput struct {
	CreditCardName string
	Year           int
	Month          time.Month
	AccountName    string
}

type StatementLine struct {
	Entry       *entity.Entry
	AccountName string
}

type CurrencyTotal struct {
	Currency string
	Amount   float64
}

type StatementOutput struct {
	Card        *entity.CreditCard
	Month       time.Time
	WindowStart time.Time
	WindowEnd   time.Time
	DueDate     time.Time
	Lines       []StatementLine
	Totals      []CurrencyTotal
}

type Statement struct {
	entryRepo   port.EntriesRepository
	ccRepo      port.CreditCardsRepository
	accountRepo port.AccountsRepository
}

func NewStatement(
	entryRepo port.EntriesRepository,
	ccRepo port.CreditCardsRepository,
	accountRepo port.AccountsRepository,
) *Statement {
	return &Statement{
		entryRepo:   entryRepo,
		ccRepo:      ccRepo,
		accountRepo: accountRepo,
	}
}

// Execute builds the invoice (fatura) for a credit card due in the given
// month: every card entry whose payment date falls within that month,
// optionally restricted to one account.
func (uc *Statement) Execute(input StatementInput) (*StatementOutput, error) {
	cc, err := uc.ccRepo.GetByName(input.CreditCardName)
	if err != nil {
		return nil, fmt.Errorf("failed to find credit card: %w", err)
	}
	if cc == nil {
		return nil, fmt.Errorf("credit card not found: %s", input.CreditCardName)
	}

	from := time.Date(input.Year, input.Month, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(input.Year, input.Month+1, 0, 0, 0, 0, 0, time.UTC)

	filters := func(accountName string) *port.EntryFilters {
		return &port.EntryFilters{
			FromDate:    &from,
			ToDate:      &to,
			CreditCard:  cc.Name,
			AccountName: accountName,
		}
	}

	var lines []StatementLine
	if input.AccountName != "" {
		entries, err := uc.entryRepo.GetAll(filters(input.AccountName))
		if err != nil {
			return nil, fmt.Errorf("failed to get entries: %w", err)
		}
		for _, entry := range entries {
			lines = append(lines, StatementLine{Entry: entry, AccountName: input.AccountName})
		}
	} else {
		accounts, err := uc.accountRepo.GetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to get accounts: %w", err)
		}
		for _, acc := range accounts {
			entries, err := uc.entryRepo.GetAll(filters(acc.Name))
			if err != nil {
				return nil, fmt.Errorf("failed to get entries: %w", err)
			}
			for _, entry := range entries {
				lines = append(lines, StatementLine{Entry: entry, AccountName: acc.Name})
			}
		}
	}

	sort.SliceStable(lines, func(i, j int) bool {
		if !lines[i].Entry.RealizationDate.Equal(lines[j].Entry.RealizationDate) {
			return lines[i].Entry.RealizationDate.Before(lines[j].Entry.RealizationDate)
		}
		return lines[i].Entry.Description < lines[j].Entry.Description
	})

	totalsByCurrency := make(map[string]float64)
	for _, line := range lines {
		totalsByCurrency[line.Entry.Currency] += line.Entry.Amount
	}

	totals := make([]CurrencyTotal, 0, len(totalsByCurrency))
	for currency, amount := range totalsByCurrency {
		totals = append(totals, CurrencyTotal{Currency: currency, Amount: amount})
	}
	sort.Slice(totals, func(i, j int) bool {
		return totals[i].Currency < totals[j].Currency
	})

	windowStart, windowEnd := cc.InvoiceWindow(input.Year, input.Month)

	return &StatementOutput{
		Card:        cc,
		Month:       from,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		DueDate:     cc.DueDate(input.Year, input.Month),
		Lines:       lines,
		Totals:      totals,
	}, nil
}
