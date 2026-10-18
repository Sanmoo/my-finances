package usecase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the Balances application module, exercised through its public
// interface with the existing yaml adapters on a temporary data path.

func setupBalancesRepos(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	err := yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"checking", "savings"},
	})
	require.NoError(t, err)

	return tempDir
}

func writeMonthEntries(t *testing.T, dataPath, account, month string, entries ...yaml.Entry) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dataPath, account, "2025"), 0755))
	require.NoError(t, yaml.Write(filepath.Join(dataPath, account, "2025", "2025-"+month+"-"+account+"-entries.yaml"), yaml.EntriesData{
		Entries: entries,
	}))
}

func writeBalancesEntries(t *testing.T, dataPath, account string, entries ...yaml.Entry) {
	t.Helper()
	writeMonthEntries(t, dataPath, account, "03", entries...)
}

func incomeEntry(amount float64, realization string, payment ...string) yaml.Entry {
	e := yaml.Entry{
		Type:            "income",
		Amount:          amount,
		Currency:        "BRL",
		Description:     "fixture",
		RealizationDate: realization,
		CreatedAt:       realization,
	}
	if len(payment) > 0 {
		e.PaymentDate = &payment[0]
	}
	return e
}

func expenseEntry(amount float64, realization string, payment ...string) yaml.Entry {
	e := incomeEntry(amount, realization, payment...)
	e.Type = "expense"
	return e
}

func date(t *testing.T, s string) *time.Time {
	t.Helper()
	d, err := time.Parse(time.DateOnly, s)
	require.NoError(t, err)
	return &d
}

func TestBalancesSelectsSingleAccountByName(t *testing.T) {
	dataPath := setupBalancesRepos(t)
	writeBalancesEntries(t, dataPath, "checking", incomeEntry(100, "2025-03-10"))
	writeBalancesEntries(t, dataPath, "savings", incomeEntry(700, "2025-03-10"))

	uc := NewBalances(yaml.NewEntriesRepository(dataPath), yaml.NewAccountsRepository(dataPath))

	out, err := uc.Execute(BalancesInput{AccountName: "savings"})
	require.NoError(t, err)

	require.Len(t, out.Accounts, 1)
	assert.Equal(t, "savings", out.Accounts[0].Name)
	assert.Equal(t, 700.0, out.Balances["savings"].TotalIncome)
	_, hasChecking := out.Balances["checking"]
	assert.False(t, hasChecking)
}

func TestBalancesAccountNotFound(t *testing.T) {
	dataPath := setupBalancesRepos(t)

	uc := NewBalances(yaml.NewEntriesRepository(dataPath), yaml.NewAccountsRepository(dataPath))

	out, err := uc.Execute(BalancesInput{AccountName: "nope"})

	assert.Nil(t, out)
	require.EqualError(t, err, "account not found: nope")
}

func TestBalancesAllAccountsInStorageOrder(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"zulu", "alpha"},
	}))
	writeMonthEntries(t, tempDir, "zulu", "03", incomeEntry(100, "2025-03-01"))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{})
	require.NoError(t, err)

	require.Len(t, out.Accounts, 2)
	assert.Equal(t, "zulu", out.Accounts[0].Name)
	assert.Equal(t, "alpha", out.Accounts[1].Name)
}

func TestBalancesAccountWithoutEntriesHasZeros(t *testing.T) {
	dataPath := setupBalancesRepos(t)

	uc := NewBalances(yaml.NewEntriesRepository(dataPath), yaml.NewAccountsRepository(dataPath))

	out, err := uc.Execute(BalancesInput{})
	require.NoError(t, err)

	for _, name := range []string{"checking", "savings"} {
		b := out.Balances[name]
		require.NotNil(t, b)
		assert.Equal(t, 0.0, b.TotalIncome)
		assert.Equal(t, 0.0, b.TotalExpense)
		assert.Equal(t, 0.0, b.Balance)
		assert.Equal(t, 0.0, b.AccumulatedBalance)
	}
}

func TestBalancesMarchAprilExample(t *testing.T) {
	// Approved example: income BRL 1000 in March, expense BRL 200 in April,
	// filter covering April: interval balance = -200, accumulated = 800.
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"checking"},
	}))
	writeMonthEntries(t, tempDir, "checking", "03", incomeEntry(1000, "2025-03-15"))
	writeMonthEntries(t, tempDir, "checking", "04", expenseEntry(200, "2025-04-20"))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{
		From:  date(t, "2025-04-01"),
		Until: date(t, "2025-04-30"),
	})
	require.NoError(t, err)

	b := out.Balances["checking"]
	assert.Equal(t, 0.0, b.TotalIncome)
	assert.Equal(t, 200.0, b.TotalExpense)
	assert.Equal(t, -200.0, b.Balance)
	assert.Equal(t, 800.0, b.AccumulatedBalance) // ignores --from
}

func TestBalancesAccumulatedRespectsUntil(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"checking"},
	}))
	writeMonthEntries(t, tempDir, "checking", "03", incomeEntry(500, "2025-03-01"))
	writeMonthEntries(t, tempDir, "checking", "06", expenseEntry(50, "2025-06-01"))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	// Until in April: the June expense is out of the accumulated cut.
	out, err := uc.Execute(BalancesInput{Until: date(t, "2025-04-30")})
	require.NoError(t, err)
	b := out.Balances["checking"]
	assert.Equal(t, 500.0, b.TotalIncome)
	assert.Equal(t, 0.0, b.TotalExpense)
	assert.Equal(t, 500.0, b.Balance)
	assert.Equal(t, 500.0, b.AccumulatedBalance)

	// Until in June: the expense enters both cuts.
	out, err = uc.Execute(BalancesInput{Until: date(t, "2025-06-30")})
	require.NoError(t, err)
	b = out.Balances["checking"]
	assert.Equal(t, 50.0, b.TotalExpense)
	assert.Equal(t, 450.0, b.Balance)
	assert.Equal(t, 450.0, b.AccumulatedBalance)

	// No Until at all: accumulated still computed over everything.
	out, err = uc.Execute(BalancesInput{})
	require.NoError(t, err)
	assert.Equal(t, 450.0, out.Balances["checking"].AccumulatedBalance)
}

func TestBalancesPaymentDateSelectsCut(t *testing.T) {
	// Realization in March, payment in April: the payment date decides.
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"card"},
	}))
	writeBalancesEntries(t, tempDir, "card", incomeEntry(100, "2025-03-15", "2025-04-10"))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{From: date(t, "2025-04-01"), Until: date(t, "2025-04-30")})
	require.NoError(t, err)
	b := out.Balances["card"]
	assert.Equal(t, 100.0, b.TotalIncome)
	assert.Equal(t, 100.0, b.AccumulatedBalance)

	out, err = uc.Execute(BalancesInput{From: date(t, "2025-03-01"), Until: date(t, "2025-03-31")})
	require.NoError(t, err)
	b = out.Balances["card"]
	assert.Equal(t, 0.0, b.TotalIncome)
	assert.Equal(t, 0.0, b.AccumulatedBalance)
}

func TestBalancesValueKinds(t *testing.T) {
	// Income, expense, negative expense (reduces TotalExpense) and raw sums
	// across currencies (legacy compatibility, no conversion).
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"checking"},
	}))
	writeBalancesEntries(t, tempDir, "checking",
		incomeEntry(1000, "2025-03-05"),
		expenseEntry(100, "2025-03-10"),
		expenseEntry(-50, "2025-03-15"),
		yaml.Entry{Type: "expense", Amount: 20, Currency: "USD", RealizationDate: "2025-03-20", CreatedAt: "2025-03-20"},
	)

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{})
	require.NoError(t, err)

	b := out.Balances["checking"]
	assert.Equal(t, 1000.0, b.TotalIncome)
	assert.Equal(t, 70.0, b.TotalExpense) // 100 - 50 + 20 (BRL+USD raw sum)
	assert.Equal(t, 930.0, b.Balance)
	assert.Equal(t, 930.0, b.AccumulatedBalance)
}

func TestBalancesGetByNameErrorPropagates(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, writeCorruptAccounts(tempDir))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{AccountName: "checking"})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal yaml")
}

func TestBalancesGetAllAccountsErrorPropagates(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, writeCorruptAccounts(tempDir))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal yaml")
}

func TestBalancesEntriesQueryErrorPropagates(t *testing.T) {
	// A corrupted accounts.yaml also fails the Entries queries (the Entries
	// adapter reads the account list first); no partial balances are returned.
	tempDir := t.TempDir()
	require.NoError(t, yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"checking"},
	}))
	writeBalancesEntries(t, tempDir, "checking", incomeEntry(100, "2025-03-10"))
	require.NoError(t, writeCorruptAccounts(tempDir))

	uc := NewBalances(yaml.NewEntriesRepository(tempDir), yaml.NewAccountsRepository(tempDir))

	out, err := uc.Execute(BalancesInput{})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal yaml")
}

func writeCorruptAccounts(dataPath string) error {
	return os.WriteFile(filepath.Join(dataPath, "accounts.yaml"), []byte("accounts: [unclosed"), 0644)
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(time.DateOnly, s)
	require.NoError(t, err)
	return d
}
