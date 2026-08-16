package usecase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupStatementRepos(t *testing.T) (*yaml.EntriesRepository, *yaml.CreditCardsRepository, *yaml.AccountsRepository) {
	t.Helper()
	tempDir := t.TempDir()

	err := yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"main", "deh"},
	})
	require.NoError(t, err)

	err = yaml.Write(filepath.Join(tempDir, "credit_cards.yaml"), yaml.CreditCardsData{
		CreditCards: []yaml.CreditCard{{Name: "main", ClosingDay: 9, DueDay: 16}},
	})
	require.NoError(t, err)

	return yaml.NewEntriesRepository(tempDir), yaml.NewCreditCardsRepository(tempDir), yaml.NewAccountsRepository(tempDir)
}

func newStatementUC(entryRepo *yaml.EntriesRepository, ccRepo *yaml.CreditCardsRepository, accountRepo *yaml.AccountsRepository) *Statement {
	return NewStatement(entryRepo, ccRepo, accountRepo)
}

func createCardEntry(t *testing.T, entryRepo *yaml.EntriesRepository, cc *entity.CreditCard, account, desc string, amount float64, currency string, realization time.Time) {
	t.Helper()
	entry, err := entity.NewEntry(
		entity.EntryTypeExpense,
		amount,
		currency,
		realization,
		entity.WithDescription(desc),
		entity.WithCreditCard(cc),
	)
	require.NoError(t, err)
	err = entryRepo.Create(entry, account)
	require.NoError(t, err)
}

func createPlainEntry(t *testing.T, entryRepo *yaml.EntriesRepository, account, desc string, amount float64, realization time.Time) {
	t.Helper()
	entry, err := entity.NewEntry(
		entity.EntryTypeExpense,
		amount,
		"BRL",
		realization,
		entity.WithDescription(desc),
	)
	require.NoError(t, err)
	err = entryRepo.Create(entry, account)
	require.NoError(t, err)
}

func TestStatement_Execute(t *testing.T) {
	t.Run("selects only entries of the card paid in the given month", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		cc, err := ccRepo.GetByName("main")
		require.NoError(t, err)

		// Payment dates with closing day 9 / due day 16:
		// realization 05-20 -> payment 06-16 (June invoice)
		// realization 06-05 -> payment 06-16 (June invoice)
		// realization 06-15 -> payment 07-16 (July invoice)
		// realization 07-10 -> payment 08-16 (August invoice)
		createCardEntry(t, entryRepo, cc, "main", "May purchase", 100, "BRL", time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "main", "June purchase", 50, "BRL", time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "main", "Late June purchase", 200, "BRL", time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "main", "July purchase", 300, "BRL", time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))
		// Non-card expense must never appear.
		createPlainEntry(t, entryRepo, "main", "Debit purchase", 10, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))

		result, err := uc.Execute(StatementInput{
			CreditCardName: "main",
			Year:           2026,
			Month:          time.June,
		})
		require.NoError(t, err)
		require.Len(t, result.Lines, 2)

		// Sorted by realization date.
		assert.Equal(t, "May purchase", result.Lines[0].Entry.Description)
		assert.Equal(t, "June purchase", result.Lines[1].Entry.Description)

		// Header fields.
		assert.Equal(t, "main", result.Card.Name)
		assert.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), result.Month)
		assert.Equal(t, time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC), result.WindowStart)
		assert.Equal(t, time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC), result.WindowEnd)
		assert.Equal(t, time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC), result.DueDate)
	})

	t.Run("filters by account", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		cc, err := ccRepo.GetByName("main")
		require.NoError(t, err)

		createCardEntry(t, entryRepo, cc, "main", "From main", 100, "BRL", time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "deh", "From deh", 200, "BRL", time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC))

		result, err := uc.Execute(StatementInput{
			CreditCardName: "main",
			Year:           2026,
			Month:          time.June,
			AccountName:    "deh",
		})
		require.NoError(t, err)
		require.Len(t, result.Lines, 1)
		assert.Equal(t, "From deh", result.Lines[0].Entry.Description)
		assert.Equal(t, "deh", result.Lines[0].AccountName)
	})

	t.Run("aggregates accounts when not filtered and tags each line", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		cc, err := ccRepo.GetByName("main")
		require.NoError(t, err)

		createCardEntry(t, entryRepo, cc, "main", "From main", 100, "BRL", time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "deh", "From deh", 200, "BRL", time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC))

		result, err := uc.Execute(StatementInput{
			CreditCardName: "main",
			Year:           2026,
			Month:          time.June,
		})
		require.NoError(t, err)
		require.Len(t, result.Lines, 2)
		assert.Equal(t, "main", result.Lines[0].AccountName)
		assert.Equal(t, "deh", result.Lines[1].AccountName)
	})

	t.Run("totals are grouped per currency and sorted", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		cc, err := ccRepo.GetByName("main")
		require.NoError(t, err)

		createCardEntry(t, entryRepo, cc, "main", "BRL one", 100, "BRL", time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "main", "BRL two", 50.5, "BRL", time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC))
		createCardEntry(t, entryRepo, cc, "main", "USD one", 20, "USD", time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC))

		result, err := uc.Execute(StatementInput{
			CreditCardName: "main",
			Year:           2026,
			Month:          time.June,
		})
		require.NoError(t, err)
		require.Len(t, result.Totals, 2)
		assert.Equal(t, "BRL", result.Totals[0].Currency)
		assert.Equal(t, 150.5, result.Totals[0].Amount)
		assert.Equal(t, "USD", result.Totals[1].Currency)
		assert.Equal(t, 20.0, result.Totals[1].Amount)
	})

	t.Run("returns empty result for a month with no charges", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		result, err := uc.Execute(StatementInput{
			CreditCardName: "main",
			Year:           2026,
			Month:          time.June,
		})
		require.NoError(t, err)
		assert.Empty(t, result.Lines)
		assert.Empty(t, result.Totals)
		assert.Equal(t, time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC), result.DueDate)
	})

	t.Run("errors when the credit card does not exist", func(t *testing.T) {
		entryRepo, ccRepo, accountRepo := setupStatementRepos(t)
		uc := newStatementUC(entryRepo, ccRepo, accountRepo)

		_, err := uc.Execute(StatementInput{
			CreditCardName: "unknown",
			Year:           2026,
			Month:          time.June,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "credit card not found")
	})
}

func TestStatement_Execute_SortsByRealizationDate(t *testing.T) {
	mockEntryRepo := new(MockEntriesRepository)
	mockCCRepo := new(MockCreditCardsRepository)
	mockAccountRepo := new(MockAccountsRepository)
	uc := NewStatement(mockEntryRepo, mockCCRepo, mockAccountRepo)

	cc := &entity.CreditCard{Name: "main", ClosingDay: 9, DueDay: 16}
	mockCCRepo.On("GetByName", "main").Return(cc, nil)

	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	mockEntryRepo.On("GetAll", mock.MatchedBy(func(f *port.EntryFilters) bool {
		return f.CreditCard == "main" &&
			f.AccountName == "deh" &&
			f.FromDate != nil && f.FromDate.Equal(from) &&
			f.ToDate != nil && f.ToDate.Equal(to)
	})).Return([]*entity.Entry{
		{Description: "June item", RealizationDate: time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC), PaymentDate: &to},
		{Description: "Zebra", RealizationDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), PaymentDate: &to},
		{Description: "Alpha", RealizationDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), PaymentDate: &to},
	}, nil)

	result, err := uc.Execute(StatementInput{
		CreditCardName: "main",
		Year:           2026,
		Month:          time.June,
		AccountName:    "deh",
	})
	require.NoError(t, err)
	require.Len(t, result.Lines, 3)

	// Same realization date -> tie broken by description.
	assert.Equal(t, "Alpha", result.Lines[0].Entry.Description)
	assert.Equal(t, "Zebra", result.Lines[1].Entry.Description)
	assert.Equal(t, "June item", result.Lines[2].Entry.Description)

	mockEntryRepo.AssertExpectations(t)
	mockCCRepo.AssertExpectations(t)
}
