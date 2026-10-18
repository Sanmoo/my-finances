package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/stretchr/testify/require"
)

// Characterization tests for `report balances`.
//
// Expectations below were extracted from the implementation at commit
// 8ecbad28b933322ac0ad6712a425eedd420e8d03 (main). They pin the observable
// behavior — stdout, error messages and formatting — so the extraction of the
// balances workflow out of the Cobra caller can be verified to preserve it.

// balancesEnv is an isolated environment: temp HOME, temp config file and a
// temp data path. Tests write YAML fixtures into the returned data path.
type balancesEnv struct {
	t        *testing.T
	dataPath string
}

func newBalancesEnv(t *testing.T, locale string) *balancesEnv {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := fmt.Sprintf("data.path: %s\nlocale: %s\ndefault.currency: BRL\n",
		filepath.Join(home, "data"), locale)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".myfin.yaml"), []byte(cfg), 0644))

	return &balancesEnv{t: t, dataPath: filepath.Join(home, "data")}
}

func (e *balancesEnv) addAccount(name string) {
	e.t.Helper()
	e.addAccounts(name)
}

func (e *balancesEnv) addAccounts(names ...string) {
	e.t.Helper()
	require.NoError(e.t, os.MkdirAll(e.dataPath, 0755))
	content := "accounts:\n"
	for _, n := range names {
		content += fmt.Sprintf("  - %s\n", n)
	}
	require.NoError(e.t, os.WriteFile(filepath.Join(e.dataPath, "accounts.yaml"), []byte(content), 0644))
}

func (e *balancesEnv) corruptAccounts() {
	e.t.Helper()
	require.NoError(e.t, os.MkdirAll(e.dataPath, 0755))
	require.NoError(e.t, os.WriteFile(filepath.Join(e.dataPath, "accounts.yaml"), []byte("accounts: [unclosed"), 0644))
}

type entryFixture struct {
	entryType       string
	amount          string
	currency        string
	realizationDate string
	paymentDate     string
}

func (e *balancesEnv) addEntry(account string, f entryFixture) {
	e.t.Helper()
	d := f.realizationDate[:10] // YYYY-MM-DD
	year, month := d[:4], d[5:7]

	yamlEntry := fmt.Sprintf("  - type: %s\n    amount: %s\n    currency: %s\n    description: fixture\n    realization_date: %q\n",
		f.entryType, f.amount, f.currency, d)
	if f.paymentDate != "" {
		yamlEntry += fmt.Sprintf("    payment_date: %q\n", f.paymentDate)
	}
	yamlEntry += "    created_at: " + fmt.Sprintf("%q", d) + "\n"

	dir := filepath.Join(e.dataPath, account, year)
	require.NoError(e.t, os.MkdirAll(dir, 0755))
	path := filepath.Join(dir, fmt.Sprintf("%s-%s-%s-entries.yaml", year, month, account))

	var content string
	if existing, err := os.ReadFile(path); err == nil {
		// Append to the month's file: the stored body already ends with a
		// newline, so the new entry joins the existing list.
		body := strings.TrimPrefix(string(existing), "entries:\n")
		content = "entries:\n" + body + yamlEntry
	} else {
		content = "entries:\n" + yamlEntry
	}
	require.NoError(e.t, os.WriteFile(path, []byte(content), 0644))
}

// runBalances resets the cached repository factory, swaps the printer output
// for a buffer, sets the flags and executes the real balances command.
// The returned string is the exact command output (the printer writes errors
// and results to the same writer).
func runBalances(t *testing.T, from, until, account, format string) string {
	t.Helper()

	var buf bytes.Buffer
	origPrinter, origRepo := printer, repo
	printer = cli.NewPrinterWithOutput(&buf)
	repo = nil
	defer func() { printer, repo = origPrinter, origRepo }()

	flags := map[string]string{"from": from, "until": until, "account": account, "format": format}
	for name, value := range flags {
		require.NoError(t, reportBalancesCmd.Flags().Set(name, value))
	}

	reportBalancesCmd.Run(reportBalancesCmd, nil)
	return buf.String()
}

const dash30 = "------------------------------"

func TestBalancesTableNoFilters(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"  Total Income:  1.000,00\n"+
		"  Total Expense: 200,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       800,00\n\n", out)
}

func TestBalancesTableIntervalAndAccumulated(t *testing.T) {
	// Approved example: income BRL 1000 in March, expense BRL 200 in April,
	// filter covering April. Interval balance = -200; accumulated up to
	// April = 800 (accumulated ignores --from).
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "2025-04-01", "2025-04-30", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"Period: from 01/04/2025 to 30/04/2025\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: 200,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       -200,00\n\n"+
		"Total Balance Up to 30/04/2025: 800,00\n\n", out)
}

func TestBalancesTableUntilOnly(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "", "2025-04-30", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"Period: Up to 30/04/2025\n\n"+
		"  Total Income:  1.000,00\n"+
		"  Total Expense: 200,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       800,00\n\n"+
		"Total Balance Up to 30/04/2025: 800,00\n\n", out)
}

func TestBalancesTableFromOnly(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "2025-03-01", "", "", "table")

	// No --until: no accumulated section is displayed.
	require.Equal(t, "=== checking ===\n\n"+
		"Period: from 01/03/2025\n\n"+
		"  Total Income:  1.000,00\n"+
		"  Total Expense: 200,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       800,00\n\n", out)
}

func TestBalancesMarkdownNoFilters(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "", "", "", "md")

	require.Equal(t, "# Balances\n\n"+
		"## checking\n\n"+
		"| | Amount |\n"+
		"|---|---:|\n"+
		"| Total Income | 1.000,00 |\n"+
		"| Total Expense | 200,00 |\n"+
		"| **Balance** | **800,00** |\n\n", out)
}

func TestBalancesMarkdownIntervalAndAccumulated(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "2025-04-01", "2025-04-30", "", "md")

	require.Equal(t, "# Balances\n\n"+
		"## checking\n\n"+
		"**Period:** from 01/04/2025 to 30/04/2025\n\n"+
		"| | Amount |\n"+
		"|---|---:|\n"+
		"| Total Income | 0,00 |\n"+
		"| Total Expense | 200,00 |\n"+
		"| **Balance** | **-200,00** |\n"+
		"| **Total Balance Up to 30/04/2025** | **800,00** |\n\n", out)
}

func TestBalancesTableLocaleEnUS(t *testing.T) {
	env := newBalancesEnv(t, "en-US")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "1000", currency: "BRL", realizationDate: "2025-03-15"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "200", currency: "BRL", realizationDate: "2025-04-20"})

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"  Total Income:  1,000.00\n"+
		"  Total Expense: 200.00\n"+
		"  "+dash30+"\n"+
		"  Balance:       800.00\n\n", out)
}

func TestBalancesEntryOutsideIntervalIgnored(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "500", currency: "BRL", realizationDate: "2025-02-10"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "50", currency: "BRL", realizationDate: "2025-06-01"})

	out := runBalances(t, "2025-03-01", "2025-04-30", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"Period: from 01/03/2025 to 30/04/2025\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       0,00\n\n"+
		"Total Balance Up to 30/04/2025: 500,00\n\n", out)
}

func TestBalancesPaymentDateSelectsMonth(t *testing.T) {
	// The filter date is the payment date when present, even when the
	// realization date falls in another month.
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("card")
	env.addEntry("card", entryFixture{entryType: "income", amount: "100", currency: "BRL", realizationDate: "2025-03-15", paymentDate: "2025-04-10"})

	// April filter: entry selected by its payment date.
	out := runBalances(t, "2025-04-01", "2025-04-30", "", "table")
	require.Equal(t, "=== card ===\n\n"+
		"Period: from 01/04/2025 to 30/04/2025\n\n"+
		"  Total Income:  100,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       100,00\n\n"+
		"Total Balance Up to 30/04/2025: 100,00\n\n", out)

	// March filter: entry excluded because its payment date is in April.
	out = runBalances(t, "2025-03-01", "2025-03-31", "", "table")
	require.Equal(t, "=== card ===\n\n"+
		"Period: from 01/03/2025 to 31/03/2025\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       0,00\n\n"+
		"Total Balance Up to 31/03/2025: 0,00\n\n", out)
}

func TestBalancesNegativeExpense(t *testing.T) {
	// A negative expense reduces TotalExpense and increases balances,
	// exactly as the current arithmetic does (legacy compatibility).
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "-50", currency: "BRL", realizationDate: "2025-03-10"})

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: -50,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       50,00\n\n", out)
}

func TestBalancesMixedCurrenciesRawSum(t *testing.T) {
	// Legacy behavior: expenses in distinct currencies are summed numerically
	// with no conversion and no currency label (100 BRL + 20 USD = 120).
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "100", currency: "BRL", realizationDate: "2025-03-10"})
	env.addEntry("checking", entryFixture{entryType: "expense", amount: "20", currency: "USD", realizationDate: "2025-03-20"})

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== checking ===\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: 120,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       -120,00\n\n", out)
}

func TestBalancesAccountWithoutEntriesShowsZeros(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("empty")

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== empty ===\n\n"+
		"  Total Income:  0,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       0,00\n\n", out)
}

func TestBalancesMarkdownEmptyAccounts(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccounts()

	out := runBalances(t, "", "", "", "md")

	require.Equal(t, "# Balances\n\n", out)
}

func TestBalancesTableNoAccounts(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccounts()

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "", out)
}

func TestBalancesMultipleAccountsPreserveOrder(t *testing.T) {
	// Accounts appear in storage order, not alphabetically.
	env := newBalancesEnv(t, "pt-BR")
	env.addAccounts("zulu", "alpha")
	env.addEntry("zulu", entryFixture{entryType: "income", amount: "100", currency: "BRL", realizationDate: "2025-03-01"})
	env.addEntry("alpha", entryFixture{entryType: "income", amount: "300", currency: "BRL", realizationDate: "2025-03-01"})

	out := runBalances(t, "", "", "", "table")

	require.Equal(t, "=== zulu ===\n\n"+
		"  Total Income:  100,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       100,00\n\n"+
		"=== alpha ===\n\n"+
		"  Total Income:  300,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       300,00\n\n", out)
}

func TestBalancesAccountFilter(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccounts("checking", "savings")
	env.addEntry("checking", entryFixture{entryType: "income", amount: "100", currency: "BRL", realizationDate: "2025-03-01"})
	env.addEntry("savings", entryFixture{entryType: "income", amount: "700", currency: "BRL", realizationDate: "2025-03-01"})

	out := runBalances(t, "", "", "savings", "table")

	require.Equal(t, "=== savings ===\n\n"+
		"  Total Income:  700,00\n"+
		"  Total Expense: 0,00\n"+
		"  "+dash30+"\n"+
		"  Balance:       700,00\n\n", out)
}

func TestBalancesAccountNotFound(t *testing.T) {
	env := newBalancesEnv(t, "pt-BR")
	env.addAccount("checking")

	out := runBalances(t, "", "", "nope", "table")

	require.Equal(t, "Error: account not found: nope\n", out)
}

func TestBalancesReadFailurePreserved(t *testing.T) {
	// A corrupted accounts.yaml fails the query; the raw storage error
	// message is printed and no balances are emitted.
	env := newBalancesEnv(t, "pt-BR")
	env.corruptAccounts()

	out := runBalances(t, "", "", "", "table")

	require.True(t, len(out) > 0 && out[:len("Error: failed to unmarshal yaml:")] == "Error: failed to unmarshal yaml:",
		"expected storage error message, got %q", out)
	require.True(t, len(out) > 0 && out[len(out)-1] == '\n')
}
