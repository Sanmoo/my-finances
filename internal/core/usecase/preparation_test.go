package usecase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Focused tests on the ReportPreparation seam — the application surface the
// report callers consume. They pin the legacy coordination semantics that the
// CLI characterization tests (cmd/report_characterization_test.go) observe
// end to end, using the real YAML adapters over temporary storage.

func setupPreparation(t *testing.T) (ReportPreparation, *yaml.EntriesRepository, port.CategoriesRepository, string) {
	t.Helper()
	tempDir := t.TempDir()

	err := yaml.Write(filepath.Join(tempDir, "accounts.yaml"), yaml.AccountsData{
		Accounts: []string{"sam", "deh"},
	})
	require.NoError(t, err)

	err = yaml.Write(filepath.Join(tempDir, "credit_cards.yaml"), yaml.CreditCardsData{
		CreditCards: []yaml.CreditCard{{Name: "card", ClosingDay: 9, DueDay: 16}},
	})
	require.NoError(t, err)

	entryRepo := yaml.NewEntriesRepository(tempDir)
	categoryRepo := yaml.NewCategoriesRepository(tempDir)
	preparation := NewReportPreparation(
		entryRepo,
		categoryRepo,
		yaml.NewAccountsRepository(tempDir),
		yaml.NewCreditCardsRepository(tempDir),
	)
	return preparation, entryRepo, categoryRepo, tempDir
}

func mustAddCategory(t *testing.T, repo port.CategoriesRepository, account, name, alias string) {
	t.Helper()
	cat, err := entity.NewCategory(name, alias, entity.CategoryTypeExpense)
	require.NoError(t, err)
	require.NoError(t, repo.Create(cat, account))
}

func mustAddEntry(t *testing.T, entryRepo *yaml.EntriesRepository, account string, entryType entity.EntryType, amount float64, realization time.Time, alias string) {
	t.Helper()
	opts := []entity.EntryOption{}
	if alias != "" {
		opts = append(opts, entity.WithCategoryAlias(alias))
	}
	entry, err := entity.NewEntry(entryType, amount, "BRL", realization, opts...)
	require.NoError(t, err)
	require.NoError(t, entryRepo.Create(entry, account))
}

func TestPrepareReportResolvesCategoriesWithLegacyAliasMap(t *testing.T) {
	preparation, entryRepo, categoryRepo, _ := setupPreparation(t)

	// sam has its own "food"; deh reuses the alias with a different name —
	// the legacy map lets the LAST account win.
	mustAddCategory(t, categoryRepo, "sam", "food", "food")
	mustAddCategory(t, categoryRepo, "deh", "Alimentação", "food")

	mustAddEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 100, time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), "food")
	mustAddEntry(t, entryRepo, "deh", entity.EntryTypeExpense, 50, time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), "food")

	result, err := preparation.PrepareReport(ReportInput{})
	require.NoError(t, err)

	require.Len(t, result.Entries, 2)
	// Entries keep their identity — no projection/rebuild.
	assert.Equal(t, 100.0, result.Entries[0].Amount)
	assert.Equal(t, 50.0, result.Entries[1].Amount)

	// Collision resolved by last account, even though both entries exist.
	food, ok := result.Categories["food"]
	require.True(t, ok)
	assert.Equal(t, "alimentação", food.Name)

	assert.Len(t, result.AccountsByName, 2)
	assert.Equal(t, 150.0, result.TotalExpense)
	assert.Equal(t, 0.0, result.TotalIncome)
}

func TestPrepareReportCategoryFilterMatchesNameOrAliasExactly(t *testing.T) {
	preparation, entryRepo, categoryRepo, _ := setupPreparation(t)

	mustAddCategory(t, categoryRepo, "sam", "transport", "transport")

	// Unresolved alias entry: dropped while a category filter is active.
	mustAddEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 10, time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), "unknown-alias")
	// No alias: also dropped.
	mustAddEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 20, time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), "")
	// Matches by alias.
	mustAddEntry(t, entryRepo, "sam", entity.EntryTypeExpense, 30, time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), "transport")

	byAlias, err := preparation.PrepareReport(ReportInput{FilterCategories: []string{"transport"}})
	require.NoError(t, err)
	require.Len(t, byAlias.Entries, 1)
	assert.Equal(t, 30.0, byAlias.Entries[0].Amount)

	// Exact equality, no case-insensitive matching.
	none, err := preparation.PrepareReport(ReportInput{FilterCategories: []string{"Transport"}})
	require.NoError(t, err)
	assert.Empty(t, none.Entries)
}

func TestPrepareReportIgnoresMissingCategoryFiles(t *testing.T) {
	preparation, entryRepo, _, _ := setupPreparation(t)

	// No categories.yaml for any account: preparation succeeds with an empty
	// map, entries keep their raw aliases.
	mustAddEntry(t, entryRepo, "sam", entity.EntryTypeIncome, 500, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), "salary")

	result, err := preparation.PrepareReport(ReportInput{})
	require.NoError(t, err)
	assert.Empty(t, result.Categories)
	require.Len(t, result.Entries, 1)
	assert.Equal(t, 500.0, result.TotalIncome)
}

func TestPrepareStatementBundlesStatementWithLegacyCategoryMap(t *testing.T) {
	preparation, entryRepo, categoryRepo, tempDir := setupPreparation(t)

	cc, err := yaml.NewCreditCardsRepository(tempDir).GetByName("card")
	require.NoError(t, err)

	emojiFood, err := entity.NewCategory("food", "food", entity.CategoryTypeExpense, entity.WithEmoji("🍔"))
	require.NoError(t, err)
	require.NoError(t, categoryRepo.Create(emojiFood, "sam"))

	// Card entry paid on 16/01 (closing 9, due 16).
	entry, err := entity.NewEntry(entity.EntryTypeExpense, 150, "BRL",
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		entity.WithDescription("Notebook"), entity.WithCreditCard(cc))
	require.NoError(t, err)
	require.NoError(t, entryRepo.Create(entry, "sam"))

	prepared, err := preparation.PrepareStatement(StatementInput{
		CreditCardName: "card",
		Year:           2026,
		Month:          time.January,
	})
	require.NoError(t, err)

	require.Len(t, prepared.Statement.Lines, 1)
	assert.Equal(t, "sam", prepared.Statement.Lines[0].AccountName)
	assert.Equal(t, "card", prepared.Statement.Card.Name)

	cat, ok := prepared.Categories["food"]
	require.True(t, ok)
	require.NotNil(t, cat.Emoji)
	assert.Equal(t, "🍔", *cat.Emoji)
	assert.Equal(t, "food", cat.Name)
}

func TestPrepareStatementPropagatesCardErrors(t *testing.T) {
	preparation, _, _, _ := setupPreparation(t)

	_, err := preparation.PrepareStatement(StatementInput{
		CreditCardName: "missing-card",
		Year:           2026,
		Month:          time.January,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credit card not found: missing-card")
}
