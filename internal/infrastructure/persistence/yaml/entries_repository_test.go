package yaml

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTempRepo(t *testing.T) (*EntriesRepository, string) {
	tempDir := t.TempDir()
	repo := NewEntriesRepository(tempDir)

	// Create accounts.yaml for GetAll tests
	accountsPath := filepath.Join(tempDir, "accounts.yaml")
	accountsData := AccountsData{Accounts: []string{"checking"}}
	err := Write(accountsPath, accountsData)
	require.NoError(t, err)

	return repo, tempDir
}

func TestEntriesRepository_Create(t *testing.T) {
	t.Run("creates entry successfully", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		entry, err := entity.NewEntry(
			entity.EntryTypeExpense,
			50.00,
			"BRL",
			date,
			entity.WithDescription("Test expense"),
		)
		require.NoError(t, err)

		err = repo.Create(entry, "checking")
		require.NoError(t, err)

		// Verify entry was saved
		entries, err := repo.GetAll(nil)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, 50.00, entries[0].Amount)
		assert.Equal(t, "Test expense", entries[0].Description)
	})

	t.Run("creates multiple entries in insertion order", func(t *testing.T) {
		repo, tempDir := setupTempRepo(t)

		// Create entries with different dates (not in chronological order)
		// First entry: March 15
		date1 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		entry1, err := entity.NewEntry(
			entity.EntryTypeExpense,
			100.00,
			"BRL",
			date1,
			entity.WithDescription("First entry - March 15"),
		)
		require.NoError(t, err)

		// Second entry: March 1 (earlier date, but added second)
		date2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
		entry2, err := entity.NewEntry(
			entity.EntryTypeExpense,
			200.00,
			"BRL",
			date2,
			entity.WithDescription("Second entry - March 1"),
		)
		require.NoError(t, err)

		// Third entry: March 20 (later date, added third)
		date3 := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
		entry3, err := entity.NewEntry(
			entity.EntryTypeExpense,
			300.00,
			"BRL",
			date3,
			entity.WithDescription("Third entry - March 20"),
		)
		require.NoError(t, err)

		// Add entries in order
		err = repo.Create(entry1, "checking")
		require.NoError(t, err)
		err = repo.Create(entry2, "checking")
		require.NoError(t, err)
		err = repo.Create(entry3, "checking")
		require.NoError(t, err)

		// Read file directly to verify insertion order is preserved
		filePath := filepath.Join(tempDir, "checking", "2024", "2024-03-checking-entries.yaml")
		data, err := Read[EntriesData](filePath)
		require.NoError(t, err)
		require.Len(t, data.Entries, 3)

		// Verify insertion order is preserved in file (not sorted)
		assert.Equal(t, "First entry - March 15", data.Entries[0].Description)
		assert.Equal(t, 100.00, data.Entries[0].Amount)
		assert.Equal(t, "Second entry - March 1", data.Entries[1].Description)
		assert.Equal(t, 200.00, data.Entries[1].Amount)
		assert.Equal(t, "Third entry - March 20", data.Entries[2].Description)
		assert.Equal(t, 300.00, data.Entries[2].Amount)

		// Verify GetAll returns sorted entries
		entries, err := repo.GetAll(nil)
		require.NoError(t, err)
		require.Len(t, entries, 3)

		// GetAll should return in chronological order
		assert.Equal(t, "Second entry - March 1", entries[0].Description)
		assert.Equal(t, "First entry - March 15", entries[1].Description)
		assert.Equal(t, "Third entry - March 20", entries[2].Description)
	})

	t.Run("creates entry with payment date (credit card)", func(t *testing.T) {
		repo, tempDir := setupTempRepo(t)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		paymentDate := time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC)
		entry, err := entity.NewEntry(
			entity.EntryTypeExpense,
			150.00,
			"BRL",
			date,
			entity.WithDescription("Credit card purchase"),
			entity.WithPaymentDate(paymentDate),
		)
		require.NoError(t, err)

		err = repo.Create(entry, "checking")
		require.NoError(t, err)

		// Entry should be stored in April file (based on payment date)
		filePath := filepath.Join(tempDir, "checking", "2024", "2024-04-checking-entries.yaml")
		_, err = os.Stat(filePath)
		require.NoError(t, err, "Entry should be stored in April file based on payment date")

		// Verify entry data
		data, err := Read[EntriesData](filePath)
		require.NoError(t, err)
		require.Len(t, data.Entries, 1)
		assert.Equal(t, "2024-03-15", data.Entries[0].RealizationDate)
		assert.Equal(t, "2024-04-10", *data.Entries[0].PaymentDate)
	})

	t.Run("creates entries in different months", func(t *testing.T) {
		repo, tempDir := setupTempRepo(t)

		// Entry in January
		date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		entry1, _ := entity.NewEntry(entity.EntryTypeExpense, 100.00, "BRL", date1)
		err := repo.Create(entry1, "checking")
		require.NoError(t, err)

		// Entry in February
		date2 := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
		entry2, _ := entity.NewEntry(entity.EntryTypeExpense, 200.00, "BRL", date2)
		err = repo.Create(entry2, "checking")
		require.NoError(t, err)

		// Entry in March
		date3 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		entry3, _ := entity.NewEntry(entity.EntryTypeExpense, 300.00, "BRL", date3)
		err = repo.Create(entry3, "checking")
		require.NoError(t, err)

		// Verify separate files were created
		janPath := filepath.Join(tempDir, "checking", "2024", "2024-01-checking-entries.yaml")
		febPath := filepath.Join(tempDir, "checking", "2024", "2024-02-checking-entries.yaml")
		marPath := filepath.Join(tempDir, "checking", "2024", "2024-03-checking-entries.yaml")

		_, err = os.Stat(janPath)
		require.NoError(t, err)
		_, err = os.Stat(febPath)
		require.NoError(t, err)
		_, err = os.Stat(marPath)
		require.NoError(t, err)

		// Verify entries are in correct files
		janData, _ := Read[EntriesData](janPath)
		assert.Len(t, janData.Entries, 1)
		assert.Equal(t, 100.00, janData.Entries[0].Amount)

		febData, _ := Read[EntriesData](febPath)
		assert.Len(t, febData.Entries, 1)
		assert.Equal(t, 200.00, febData.Entries[0].Amount)

		marData, _ := Read[EntriesData](marPath)
		assert.Len(t, marData.Entries, 1)
		assert.Equal(t, 300.00, marData.Entries[0].Amount)
	})
}

func TestEntriesRepository_Create_PreservesInsertionOrder(t *testing.T) {
	repo, tempDir := setupTempRepo(t)

	// Create entries in specific insertion order with mixed dates
	entries := []struct {
		date        time.Time
		amount      float64
		description string
	}{
		{time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC), 10.00, "Entry A - March 25"},
		{time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC), 20.00, "Entry B - March 5"},
		{time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), 30.00, "Entry C - March 15"},
		{time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC), 40.00, "Entry D - March 30"},
		{time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), 50.00, "Entry E - March 1"},
	}

	// Insert entries
	for _, e := range entries {
		entry, err := entity.NewEntry(
			entity.EntryTypeExpense,
			e.amount,
			"BRL",
			e.date,
			entity.WithDescription(e.description),
		)
		require.NoError(t, err)
		err = repo.Create(entry, "checking")
		require.NoError(t, err)
	}

	// Read file directly - should preserve insertion order
	filePath := filepath.Join(tempDir, "checking", "2024", "2024-03-checking-entries.yaml")
	data, err := Read[EntriesData](filePath)
	require.NoError(t, err)
	require.Len(t, data.Entries, 5)

	// Verify insertion order (A, B, C, D, E) - not sorted
	expectedOrder := []string{
		"Entry A - March 25",
		"Entry B - March 5",
		"Entry C - March 15",
		"Entry D - March 30",
		"Entry E - March 1",
	}

	for i, expected := range expectedOrder {
		assert.Equal(t, expected, data.Entries[i].Description,
			"Entry at position %d should be '%s'", i, expected)
	}

	// Verify GetAll returns chronological order
	allEntries, err := repo.GetAll(nil)
	require.NoError(t, err)
	require.Len(t, allEntries, 5)

	// Should be sorted: E (Mar 1), B (Mar 5), C (Mar 15), A (Mar 25), D (Mar 30)
	expectedChronological := []string{
		"Entry E - March 1",
		"Entry B - March 5",
		"Entry C - March 15",
		"Entry A - March 25",
		"Entry D - March 30",
	}

	for i, expected := range expectedChronological {
		assert.Equal(t, expected, allEntries[i].Description,
			"Entry at position %d should be '%s' in sorted results", i, expected)
	}
}

func TestEntriesRepository_Create_WithPaymentDates(t *testing.T) {
	repo, tempDir := setupTempRepo(t)

	// Create entries with different payment dates spanning multiple months
	entries := []struct {
		realizationDate time.Time
		paymentDate     time.Time
		description     string
		amount          float64
	}{
		{
			realizationDate: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
			paymentDate:     time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC),
			description:     "Entry 1 - realized Mar, paid Apr",
			amount:          100.00,
		},
		{
			realizationDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			paymentDate:     time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC), // Same payment date
			description:     "Entry 2 - realized Mar, paid Apr",
			amount:          200.00,
		},
		{
			realizationDate: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			paymentDate:     time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC), // Different month
			description:     "Entry 3 - realized Mar, paid May",
			amount:          300.00,
		},
	}

	// Insert entries
	for _, e := range entries {
		entry, err := entity.NewEntry(
			entity.EntryTypeExpense,
			e.amount,
			"BRL",
			e.realizationDate,
			entity.WithDescription(e.description),
			entity.WithPaymentDate(e.paymentDate),
		)
		require.NoError(t, err)
		err = repo.Create(entry, "checking")
		require.NoError(t, err)
	}

	// Entry 1 and 2 should be in April file
	aprPath := filepath.Join(tempDir, "checking", "2024", "2024-04-checking-entries.yaml")
	aprData, err := Read[EntriesData](aprPath)
	require.NoError(t, err)
	require.Len(t, aprData.Entries, 2)

	// Should be in insertion order
	assert.Equal(t, "Entry 1 - realized Mar, paid Apr", aprData.Entries[0].Description)
	assert.Equal(t, "Entry 2 - realized Mar, paid Apr", aprData.Entries[1].Description)

	// Entry 3 should be in May file
	mayPath := filepath.Join(tempDir, "checking", "2024", "2024-05-checking-entries.yaml")
	mayData, err := Read[EntriesData](mayPath)
	require.NoError(t, err)
	require.Len(t, mayData.Entries, 1)
	assert.Equal(t, "Entry 3 - realized Mar, paid May", mayData.Entries[0].Description)

	// Verify GetAll sorts by payment date
	allEntries, err := repo.GetAll(nil)
	require.NoError(t, err)
	require.Len(t, allEntries, 3)

	// First two have same payment date (Apr 5), so sorted by realization date
	assert.Equal(t, "Entry 1 - realized Mar, paid Apr", allEntries[0].Description)
	assert.Equal(t, "Entry 2 - realized Mar, paid Apr", allEntries[1].Description)
	assert.Equal(t, "Entry 3 - realized Mar, paid May", allEntries[2].Description)
}

func TestEntriesRepository_GetAll(t *testing.T) {
	t.Run("returns all entries sorted by date", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		// Create entries with different dates
		dates := []time.Time{
			time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		}

		for i, date := range dates {
			entry, err := entity.NewEntry(
				entity.EntryTypeExpense,
				float64(100*(i+1)),
				"BRL",
				date,
			)
			require.NoError(t, err)
			err = repo.Create(entry, "checking")
			require.NoError(t, err)
		}

		// GetAll should return entries sorted
		entries, err := repo.GetAll(nil)
		require.NoError(t, err)
		require.Len(t, entries, 3)

		// Should be sorted: Mar 5, Mar 15, Mar 25
		assert.Equal(t, 5, entries[0].RealizationDate.Day())
		assert.Equal(t, 15, entries[1].RealizationDate.Day())
		assert.Equal(t, 25, entries[2].RealizationDate.Day())
	})

	t.Run("filters by type", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		// Create expense
		expense, _ := entity.NewEntry(entity.EntryTypeExpense, 100.00, "BRL", date)
		err := repo.Create(expense, "checking")
		require.NoError(t, err)

		// Create income
		income, _ := entity.NewEntry(entity.EntryTypeIncome, 500.00, "BRL", date)
		err = repo.Create(income, "checking")
		require.NoError(t, err)

		// Filter by expense type
		expenseType := entity.EntryTypeExpense
		filters := &port.EntryFilters{Type: &expenseType}
		entries, err := repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, entity.EntryTypeExpense, entries[0].Type)

		// Filter by income type
		incomeType := entity.EntryTypeIncome
		filters = &port.EntryFilters{Type: &incomeType}
		entries, err = repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, entity.EntryTypeIncome, entries[0].Type)
	})

	t.Run("filters by date range", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		// Create entries on different dates
		entries := []struct {
			date   time.Time
			amount float64
		}{
			{time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), 100.00},
			{time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), 200.00},
			{time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), 300.00},
			{time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC), 400.00},
		}

		for _, e := range entries {
			entry, _ := entity.NewEntry(entity.EntryTypeExpense, e.amount, "BRL", e.date)
			err := repo.Create(entry, "checking")
			require.NoError(t, err)
		}

		// Filter from March 10 to March 25
		fromDate := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
		filters := &port.EntryFilters{
			FromDate: &fromDate,
			ToDate:   &toDate,
		}

		result, err := repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, 200.00, result[0].Amount) // March 10
		assert.Equal(t, 300.00, result[1].Amount) // March 20
	})

	t.Run("filters by category", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		// Create entry with category
		entry1, _ := entity.NewEntry(
			entity.EntryTypeExpense,
			100.00,
			"BRL",
			date,
			entity.WithCategoryAlias("food"),
		)
		err := repo.Create(entry1, "checking")
		require.NoError(t, err)

		// Create entry without category
		entry2, _ := entity.NewEntry(entity.EntryTypeExpense, 200.00, "BRL", date)
		err = repo.Create(entry2, "checking")
		require.NoError(t, err)

		// Create entry with different category
		entry3, _ := entity.NewEntry(
			entity.EntryTypeExpense,
			300.00,
			"BRL",
			date,
			entity.WithCategoryAlias("transport"),
		)
		err = repo.Create(entry3, "checking")
		require.NoError(t, err)

		// Filter by food category
		filters := &port.EntryFilters{CategoryAlias: "food"}
		entries, err := repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, 100.00, entries[0].Amount)
	})

	t.Run("filters by tags", func(t *testing.T) {
		repo, _ := setupTempRepo(t)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		// Create entry with tags
		entry1, _ := entity.NewEntry(
			entity.EntryTypeExpense,
			100.00,
			"BRL",
			date,
			entity.WithTags([]string{"monthly", "subscription"}),
		)
		err := repo.Create(entry1, "checking")
		require.NoError(t, err)

		// Create entry without tags
		entry2, _ := entity.NewEntry(entity.EntryTypeExpense, 200.00, "BRL", date)
		err = repo.Create(entry2, "checking")
		require.NoError(t, err)

		// Create entry with different tags
		entry3, _ := entity.NewEntry(
			entity.EntryTypeExpense,
			300.00,
			"BRL",
			date,
			entity.WithTags([]string{"one-time"}),
		)
		err = repo.Create(entry3, "checking")
		require.NoError(t, err)

		// Filter by tag
		filters := &port.EntryFilters{Tags: []string{"monthly"}}
		entries, err := repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, 100.00, entries[0].Amount)
	})

	t.Run("filters by account", func(t *testing.T) {
		tempDir := t.TempDir()
		repo := NewEntriesRepository(tempDir)

		// Setup accounts.yaml with multiple accounts
		accountsPath := filepath.Join(tempDir, "accounts.yaml")
		accountsData := AccountsData{Accounts: []string{"checking", "savings"}}
		err := Write(accountsPath, accountsData)
		require.NoError(t, err)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		// Create entry in checking
		entry1, _ := entity.NewEntry(entity.EntryTypeExpense, 100.00, "BRL", date)
		err = repo.Create(entry1, "checking")
		require.NoError(t, err)

		// Create entry in savings
		entry2, _ := entity.NewEntry(entity.EntryTypeExpense, 200.00, "BRL", date)
		err = repo.Create(entry2, "savings")
		require.NoError(t, err)

		// Filter by checking account
		filters := &port.EntryFilters{AccountName: "checking"}
		entries, err := repo.GetAll(filters)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, 100.00, entries[0].Amount)
	})
}

func TestEntriesRepository_GetAll_Empty(t *testing.T) {
	repo, _ := setupTempRepo(t)

	entries, err := repo.GetAll(nil)
	require.NoError(t, err)
	assert.Empty(t, entries)
}
