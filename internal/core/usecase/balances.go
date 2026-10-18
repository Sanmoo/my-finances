package usecase

import (
	"errors"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

// Balance is the application-owned financial result for one Account: the
// sums for the selected interval, the interval balance and the accumulated
// balance up to the final limit.
//
// Values keep the legacy float64 arithmetic and the raw numeric sum across
// distinct currencies; no rounding, minor units or currency conversion is
// applied here. That mixing is a compatibility contract, not a domain rule.
type Balance struct {
	Account            *entity.Account
	TotalIncome        float64
	TotalExpense       float64
	Balance            float64
	AccumulatedBalance float64
}

type BalancesInput struct {
	// AccountName restricts the workflow to a single Account. When empty,
	// all Accounts are used in repository order (not alphabetically), and
	// Accounts without Entries appear with zero values.
	AccountName string
	// From and Until bound the interval cut. From is ignored by the
	// accumulated cut, which considers everything up to Until.
	From  *time.Time
	Until *time.Time
}

type BalancesOutput struct {
	Accounts []*entity.Account
	Balances map[string]*Balance
}

// Balances is the application module behind `report balances`. It owns the
// whole workflow: Account selection/validation, the two Entries queries
// (interval and accumulated), the arithmetic and the assembly of financial
// results. Callers keep only flags/dates, adapter wiring, execution and the
// choice of formatting/output.
//
// The filter date of an Entry is its Payment date when present, otherwise
// its Realization date; boundaries and timezone comparison follow the
// EntriesRepository contract unchanged.
type Balances struct {
	entryRepo   port.EntriesRepository
	accountRepo port.AccountsRepository
}

func NewBalances(entryRepo port.EntriesRepository, accountRepo port.AccountsRepository) *Balances {
	return &Balances{
		entryRepo:   entryRepo,
		accountRepo: accountRepo,
	}
}

func (uc *Balances) Execute(input BalancesInput) (*BalancesOutput, error) {
	var accounts []*entity.Account

	if input.AccountName != "" {
		account, err := uc.accountRepo.GetByName(input.AccountName)
		if err != nil {
			return nil, err
		}
		if account == nil {
			return nil, errors.New("account not found: " + input.AccountName)
		}
		accounts = []*entity.Account{account}
	} else {
		var err error
		accounts, err = uc.accountRepo.GetAll()
		if err != nil {
			return nil, err
		}
	}

	balances := make(map[string]*Balance)

	for _, acc := range accounts {
		// Entries within the selected interval (respecting From and Until).
		entries, err := uc.entryRepo.GetAll(&port.EntryFilters{
			FromDate:    input.From,
			ToDate:      input.Until,
			AccountName: acc.Name,
		})
		if err != nil {
			return nil, err
		}

		var totalIncome, totalExpense float64
		for _, entry := range entries {
			if entry.IsIncome() {
				totalIncome += entry.Amount
			} else {
				totalExpense += entry.Amount
			}
		}

		// Accumulated entries up to Until (ignoring From).
		accumulatedEntries, err := uc.entryRepo.GetAll(&port.EntryFilters{
			ToDate:      input.Until,
			AccountName: acc.Name,
		})
		if err != nil {
			return nil, err
		}

		var accumulatedBalance float64
		for _, entry := range accumulatedEntries {
			if entry.IsIncome() {
				accumulatedBalance += entry.Amount
			} else {
				accumulatedBalance -= entry.Amount
			}
		}

		balances[acc.Name] = &Balance{
			Account:            acc,
			TotalIncome:        totalIncome,
			TotalExpense:       totalExpense,
			Balance:            totalIncome - totalExpense,
			AccumulatedBalance: accumulatedBalance,
		}
	}

	return &BalancesOutput{Accounts: accounts, Balances: balances}, nil
}
