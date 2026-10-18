package main

// Characterization tests for `report entries`, `report by-category` and
// `report statement`.
//
// These tests pin the CURRENT observable behavior of the three commands,
// including legacy quirks that are deliberately NOT fixed here (see issue #3):
//
//   - Categories are resolved through a global map keyed by alias, built by
//     iterating ALL accounts (in accounts.yaml order, even when --account is
//     given) and letting the LAST account's category win on alias collisions.
//   - Category read errors are silently ignored; a missing categories file is
//     just an empty list.
//   - by-category sums amounts across different currencies and labels the
//     totals as BRL; groups are merged by final display text (emoji included).
//   - Unresolved category aliases are shown raw in by-category and as an empty
//     category in entries; entries without an alias are grouped as
//     "Sem categoria".
//   - Entries/statement display the Payment date when present, even with
//     --by-realization (that flag only changes the filtering date).
//   - by-category group order for equal totals is unspecified (sort.Slice, not
//     stable) — the fixtures below avoid ties so byte-exact goldens are valid.
//
// The goldens in testdata/characterization were produced by the reference
// implementation (8ecbad2) and must only be regenerated when the reference
// behavior itself changes — never to normalize a refactoring divergence.

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
)

var updateGoldens = flag.Bool("update-goldens", false, "regenerate characterization goldens from the current implementation")

// runCommand replaces the package-level printer/repo wiring with a buffered
// printer and a repository factory over an isolated data dir, then executes
// the real Cobra command tree. It returns everything the command printed.
func runCommand(t *testing.T, dataDir string, args ...string) string {
	t.Helper()

	// pflag values persist across Execute() calls on the same command; reset
	// them so each invocation starts from the flag defaults, like a fresh CLI.
	resetReportFlags(t)

	var buf bytes.Buffer
	oldPrinter, oldRepo := printer, repo
	printer = cli.NewPrinterWithOutput(&buf)
	repo = persistence.NewRepositoryFactory(&config.Config{DataPath: dataDir})
	t.Cleanup(func() { printer, repo = oldPrinter, oldRepo })

	rootCmd.SetArgs(args)
	require.NoError(t, rootCmd.Execute())
	return buf.String()
}

func resetReportFlags(t *testing.T) {
	t.Helper()
	for _, c := range []*cobra.Command{reportEntriesCmd, reportByCategoryCmd, reportStatementCmd} {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			f.Changed = false
			require.NoError(t, f.Value.Set(f.DefValue))
		})
	}
}

func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "characterization", name+".golden")
	if *updateGoldens {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(got), 0644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "missing golden file %s; run tests with -update-goldens to generate", path)
	assert.Equal(t, string(want), got, "output of golden %q diverged from the reference implementation", name)
}

// writeFixture builds the shared deterministic storage:
//
// Accounts: sam, deh (in this order — the alias map iteration order).
// Card:     card-sam (closing 9, due 16) used by both accounts.
// Categories (loaded in account order, later accounts overwrite aliases):
//
//	sam: food  -> "food" (🍔)      display "🍔 food"
//	     sam-food -> "food" (🍔)   display "🍔 food"
//	     transport -> "Transport"  display "Transport"
//	deh: food  -> "Alimentação"    display "Alimentação"  (wins over sam's food)
//	     deh-food -> "food" (🍔)   display "🍔 food"
//
// Entries (distinct dates, no ordering ties):
//
//	sam:  2026-01-05 income  1000.00 BRL "Salary"     alias salary (unresolved)
//	      2026-01-10 expense  123.45 BRL "Supermarket" alias food -> Alimentação, tag market
//	      2026-01-11 expense  -20.00 BRL "Refund"      no alias
//	      2026-01-12 expense   30.00 BRL "Unresolved"  alias unknown-alias (unresolved)
//	      2026-01-02 expense  150.00 BRL "Notebook"    card-sam 1/3 (payment 16/01)
//	deh:  2026-01-20 expense  200.00 BRL "Delivery"    alias sam-food -> 🍔 food
//	      2026-01-21 expense   80.50 BRL "Lunch"       alias deh-food -> 🍔 food (same text as above)
//	      2026-01-22 expense   60.00 BRL "Deh misc"    no alias, tag trip
//	      2026-01-05 expense  300.00 USD "Card dinner" card-sam (payment 16/01)
func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, yaml.Write(filepath.Join(dir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"sam", "deh"},
	}))
	require.NoError(t, yaml.Write(filepath.Join(dir, "credit_cards.yaml"), yaml.CreditCardsData{
		CreditCards: []yaml.CreditCard{{Name: "card-sam", ClosingDay: 9, DueDay: 16}},
	}))

	catRepo := yaml.NewCategoriesRepository(dir)
	createCategory(t, catRepo, "food", "food", entity.CategoryTypeExpense, "🍔", "sam")
	createCategory(t, catRepo, "food", "sam-food", entity.CategoryTypeExpense, "🍔", "sam")
	createCategory(t, catRepo, "transport", "transport", entity.CategoryTypeExpense, "", "sam")
	createCategory(t, catRepo, "Alimentação", "food", entity.CategoryTypeExpense, "", "deh")
	createCategory(t, catRepo, "food", "deh-food", entity.CategoryTypeExpense, "🍔", "deh")

	entryRepo := yaml.NewEntriesRepository(dir)
	cc, err := yaml.NewCreditCardsRepository(dir).GetByName("card-sam")
	require.NoError(t, err)
	require.NotNil(t, cc)

	d := func(day int) time.Time { return time.Date(2026, time.January, day, 0, 0, 0, 0, time.UTC) }

	createEntry(t, entryRepo, "sam", entity.EntryTypeIncome, 1000, "BRL", d(5), "Salary", "salary", nil, nil, 0, 0)
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 123.45, "BRL", d(10), "Supermarket", "food", []string{"market"}, nil, 0, 0)
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, -20, "BRL", d(11), "Refund", "", nil, nil, 0, 0)
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 30, "BRL", d(12), "Unresolved", "unknown-alias", nil, nil, 0, 0)
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 150, "BRL", d(2), "Notebook", "", nil, cc, 1, 3)
	createEntry(t, entryRepo, "deh", entity.EntryTypeExpense, 200, "BRL", d(20), "Delivery", "sam-food", nil, nil, 0, 0)
	createEntry(t, entryRepo, "deh", entity.EntryTypeExpense, 80.5, "BRL", d(21), "Lunch", "deh-food", nil, nil, 0, 0)
	createEntry(t, entryRepo, "deh", entity.EntryTypeExpense, 60, "BRL", d(22), "Deh misc", "", []string{"trip"}, nil, 0, 0)
	createEntry(t, entryRepo, "deh", entity.EntryTypeExpense, 300, "USD", d(5), "Card dinner", "", nil, cc, 0, 0)

	return dir
}

func createCategory(t *testing.T, repo *yaml.CategoriesRepository, name, alias string, typ entity.CategoryType, emoji, account string) {
	t.Helper()
	opts := []entity.CategoryOption{}
	if emoji != "" {
		opts = append(opts, entity.WithEmoji(emoji))
	}
	cat, err := entity.NewCategory(name, alias, typ, opts...)
	require.NoError(t, err)
	require.NoError(t, repo.Create(cat, account))
}

func createEntry(t *testing.T, repo *yaml.EntriesRepository, account string, entryType entity.EntryType, amount float64, currency string, realization time.Time, desc, catAlias string, tags []string, card *entity.CreditCard, instNum, instTotal int) {
	t.Helper()
	opts := []entity.EntryOption{}
	if desc != "" {
		opts = append(opts, entity.WithDescription(desc))
	}
	if catAlias != "" {
		opts = append(opts, entity.WithCategoryAlias(catAlias))
	}
	if len(tags) > 0 {
		opts = append(opts, entity.WithTags(tags))
	}
	if instTotal > 0 {
		opts = append(opts, entity.WithInstallment(instNum, instTotal))
	}
	if card != nil {
		opts = append(opts, entity.WithCreditCard(card))
	}
	entry, err := entity.NewEntry(entryType, amount, currency, realization, opts...)
	require.NoError(t, err)
	require.NoError(t, repo.Create(entry, account))
}

// --- report entries -------------------------------------------------------

func TestEntriesTableDefault(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "entries")
	compareGolden(t, "entries_table_default", got)
}

func TestEntriesMarkdownDefault(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "entries", "--format", "md")
	compareGolden(t, "entries_markdown_default", got)
}

func TestEntriesTableAccountFilter(t *testing.T) {
	dir := writeFixture(t)
	// Category resolution still sees ALL accounts (legacy): deh's alias
	// "food" -> "Alimentação" wins even though only deh entries are listed.
	got := runCommand(t, dir, "report", "entries", "--account", "deh")
	compareGolden(t, "entries_table_account_filter", got)
}

func TestEntriesTableDateFilters(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "entries", "--from", "2026-01-10", "--until", "2026-01-20")
	compareGolden(t, "entries_table_date_filters", got)
}

func TestEntriesTableTagFilter(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "entries", "--filter-tags", "market")
	compareGolden(t, "entries_table_tag_filter", got)
}

func TestEntriesTableCategoryFilterByName(t *testing.T) {
	dir := writeFixture(t)
	// Filter matches the resolved category NAME (stored lowercased). Both the
	// "alimentação" match and the ones via alias "food" pass the filter.
	got := runCommand(t, dir, "report", "entries", "--filter-categories", "alimentação")
	compareGolden(t, "entries_table_category_filter_by_name", got)
}

func TestEntriesTableCategoryFilterCaseSensitive(t *testing.T) {
	dir := writeFixture(t)
	// Matching is exact equality: no category is stored as "Alimentação"
	// (names are lowercased), so nothing matches.
	got := runCommand(t, dir, "report", "entries", "--filter-categories", "Alimentação")
	compareGolden(t, "entries_table_category_filter_case_sensitive", got)
}

func TestEntriesTableCategoryFilterByAlias(t *testing.T) {
	dir := writeFixture(t)
	// Filter matches the resolved category ALIAS; entries whose alias does not
	// resolve are excluded while the filter is active.
	got := runCommand(t, dir, "report", "entries", "--filter-categories", "food")
	compareGolden(t, "entries_table_category_filter_by_alias", got)
}

func TestEntriesTableByRealizationChangesFilterNotDisplay(t *testing.T) {
	dir := writeFixture(t)
	// The Notebook installment has realization 02/01 and payment 16/01.
	// Default filtering is payment-based, so --from 2026-01-03 keeps it and
	// displays 16/01; --by-realization filters by 02/01 and drops it, without
	// changing the displayed date of the remaining rows.
	paymentBased := runCommand(t, dir, "report", "entries", "--from", "2026-01-03")
	realizationBased := runCommand(t, dir, "report", "entries", "--from", "2026-01-03", "--by-realization")
	compareGolden(t, "entries_table_from_payment_based", paymentBased)
	compareGolden(t, "entries_table_from_realization_based", realizationBased)
}

func TestEntriesTableNoResults(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "entries", "--filter-tags", "nonexistent-tag")
	compareGolden(t, "entries_table_no_results", got)
}

// --- report by-category ---------------------------------------------------

func TestByCategoryTableDefault(t *testing.T) {
	dir := writeFixture(t)
	// Includes: alias collision resolution, merged "🍔 food" display text,
	// "Sem categoria" group, raw unresolved aliases, cross-currency numeric
	// sum with BRL label and the extra table subtotal.
	got := runCommand(t, dir, "report", "by-category")
	compareGolden(t, "by_category_table_default", got)
}

func TestByCategoryMarkdownDefault(t *testing.T) {
	dir := writeFixture(t)
	// Markdown has no subtotal line.
	got := runCommand(t, dir, "report", "by-category", "--format", "md")
	compareGolden(t, "by_category_markdown_default", got)
}

func TestByCategoryTableAccountFilter(t *testing.T) {
	dir := writeFixture(t)
	// Same legacy category map (all accounts), only deh entries grouped.
	got := runCommand(t, dir, "report", "by-category", "--account", "deh")
	compareGolden(t, "by_category_table_account_filter", got)
}

// --- missing / unreadable category files ---------------------------------

// writeFixtureWithoutCategories stores the same entries but no categories.yaml
// file at all: the legacy behavior treats it as an empty category list, so
// every alias is unresolved and no failure is reported.
func writeFixtureWithoutCategories(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, yaml.Write(filepath.Join(dir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"sam"},
	}))

	entryRepo := yaml.NewEntriesRepository(dir)
	d := func(day int) time.Time { return time.Date(2026, time.January, day, 0, 0, 0, 0, time.UTC) }
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 123.45, "BRL", d(10), "Supermarket", "food", nil, nil, 0, 0)
	createEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 60, "BRL", d(11), "No category", "", nil, nil, 0, 0)
	return dir
}

func TestEntriesTableMissingCategoriesFile(t *testing.T) {
	dir := writeFixtureWithoutCategories(t)
	got := runCommand(t, dir, "report", "entries")
	compareGolden(t, "entries_table_missing_categories_file", got)
}

func TestByCategoryTableMissingCategoriesFile(t *testing.T) {
	dir := writeFixtureWithoutCategories(t)
	got := runCommand(t, dir, "report", "by-category")
	compareGolden(t, "by_category_table_missing_categories_file", got)
}

func TestByCategoryTableUnreadableCategoriesFile(t *testing.T) {
	dir := writeFixture(t)
	// A malformed categories.yaml makes GetAll fail; the legacy flows ignore
	// category read errors (no warning, no aggregated error) and treat the
	// account's categories as absent.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "deh", "categories.yaml"),
		[]byte("categories: [not: valid: yaml"), 0644))
	got := runCommand(t, dir, "report", "by-category")
	compareGolden(t, "by_category_table_unreadable_categories_file", got)
}

// --- report statement -----------------------------------------------------

func TestStatementTableMultipleAccounts(t *testing.T) {
	dir := writeFixture(t)
	// Lines from two accounts -> Account column shown; BRL and USD totals
	// reported separately.
	got := runCommand(t, dir, "report", "statement", "--card", "card-sam", "--month", "2026-01")
	compareGolden(t, "statement_table_multiple_accounts", got)
}

func TestStatementMarkdownMultipleAccounts(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "statement", "--card", "card-sam", "--month", "2026-01", "--format", "md")
	compareGolden(t, "statement_markdown_multiple_accounts", got)
}

func TestStatementTableSingleAccount(t *testing.T) {
	dir := writeFixture(t)
	// Single account -> no Account column.
	got := runCommand(t, dir, "report", "statement", "--card", "card-sam", "--month", "2026-01", "--account", "sam")
	compareGolden(t, "statement_table_single_account", got)
}

func TestStatementTableEmptyInvoice(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "statement", "--card", "card-sam", "--month", "2026-03")
	compareGolden(t, "statement_table_empty_invoice", got)
}

func TestStatementNonexistentCard(t *testing.T) {
	dir := writeFixture(t)
	got := runCommand(t, dir, "report", "statement", "--card", "missing-card", "--month", "2026-01")
	compareGolden(t, "statement_nonexistent_card", got)
}
