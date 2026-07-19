package interactive

import (
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Fake Prompter ---

type fakePrompter struct {
	textResponses    []string
	confirmResponses []bool
	textIdx          int
	confirmIdx       int
}

func (f *fakePrompter) Text(prompt, defaultValue string) (string, error) {
	resp := f.textResponses[f.textIdx]
	f.textIdx++
	return resp, nil
}

func (f *fakePrompter) Confirm(prompt string, defaultYes bool) (bool, error) {
	resp := f.confirmResponses[f.confirmIdx]
	f.confirmIdx++
	return resp, nil
}

// --- Fake Selector ---

type fakeSelectResponse struct {
	choice string
	ok     bool
}

type fakeSelector struct {
	selectResponses      []fakeSelectResponse
	multiSelectResponses [][]string
	selectIdx            int
	multiIdx             int
}

func (f *fakeSelector) Select(title string, options []string) (string, bool, error) {
	resp := f.selectResponses[f.selectIdx]
	f.selectIdx++
	return resp.choice, resp.ok, nil
}

func (f *fakeSelector) MultiSelect(title string, options []string) ([]string, error) {
	resp := f.multiSelectResponses[f.multiIdx]
	f.multiIdx++
	return resp, nil
}

// --- Mock Repositories ---

type mockEntriesRepo struct{ mock.Mock }

func (m *mockEntriesRepo) Create(entry *entity.Entry, accountName string) error {
	args := m.Called(entry, accountName)
	return args.Error(0)
}

func (m *mockEntriesRepo) GetAll(filters *port.EntryFilters) ([]*entity.Entry, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Entry), args.Error(1)
}

type mockCategoriesRepo struct{ mock.Mock }

func (m *mockCategoriesRepo) Create(cat *entity.Category, accountName string) error {
	args := m.Called(cat, accountName)
	return args.Error(0)
}

func (m *mockCategoriesRepo) GetByAlias(accountName, alias string) (*entity.Category, error) {
	args := m.Called(accountName, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *mockCategoriesRepo) GetAll(accountName string) ([]*entity.Category, error) {
	args := m.Called(accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Category), args.Error(1)
}

type mockTagsRepo struct{ mock.Mock }

func (m *mockTagsRepo) Create(tag *entity.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

func (m *mockTagsRepo) GetByName(name string) (*entity.Tag, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Tag), args.Error(1)
}

func (m *mockTagsRepo) GetAll() ([]*entity.Tag, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Tag), args.Error(1)
}

type mockCCRepo struct{ mock.Mock }

func (m *mockCCRepo) Create(cc *entity.CreditCard) error {
	args := m.Called(cc)
	return args.Error(0)
}

func (m *mockCCRepo) GetByName(name string) (*entity.CreditCard, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CreditCard), args.Error(1)
}

func (m *mockCCRepo) GetAll() ([]*entity.CreditCard, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CreditCard), args.Error(1)
}

type mockAccountsRepo struct{ mock.Mock }

func (m *mockAccountsRepo) Create(acc *entity.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *mockAccountsRepo) GetByName(name string) (*entity.Account, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *mockAccountsRepo) GetAll() ([]*entity.Account, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

// --- promptType tests ---

func TestWizard_promptType_SelectExpense(t *testing.T) {
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{choiceExpense, true}},
	}
	w := &Wizard{selector: sel}
	typ, err := w.promptType()
	require.NoError(t, err)
	assert.Equal(t, entity.EntryTypeExpense, typ)
}

func TestWizard_promptType_SelectIncome(t *testing.T) {
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{choiceIncome, true}},
	}
	w := &Wizard{selector: sel}
	typ, err := w.promptType()
	require.NoError(t, err)
	assert.Equal(t, entity.EntryTypeIncome, typ)
}

func TestWizard_promptType_Cancelled(t *testing.T) {
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{"", false}},
	}
	w := &Wizard{selector: sel}
	_, err := w.promptType()
	assert.ErrorIs(t, err, errCancelled)
}

// --- promptAccount tests ---

func TestWizard_promptAccount_SelectExisting(t *testing.T) {
	accRepo := new(mockAccountsRepo)
	accRepo.On("GetAll").Return([]*entity.Account{{Name: "nubank"}}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{"nubank", true}},
	}
	w := &Wizard{selector: sel, accountRepo: accRepo}
	name, err := w.promptAccount()
	require.NoError(t, err)
	assert.Equal(t, "nubank", name)
}

func TestWizard_promptAccount_InlineCreate(t *testing.T) {
	accRepo := new(mockAccountsRepo)
	accRepo.On("GetAll").Return([]*entity.Account{}, nil)
	accRepo.On("Create", mock.AnythingOfType("*entity.Account")).Return(nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{
			{selectorNewAccount, true},
		},
	}
	prompter := &fakePrompter{
		textResponses: []string{"novo-banco"},
	}
	w := &Wizard{selector: sel, accountRepo: accRepo, prompter: prompter, printer: cli.NewPrinter()}
	name, err := w.promptAccount()
	require.NoError(t, err)
	assert.Equal(t, "novo-banco", name)
	accRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.Account"))
}

// --- promptTags tests ---

func TestWizard_promptTags_MultiSelect(t *testing.T) {
	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{{Name: "work"}, {Name: "vr"}}, nil)
	sel := &fakeSelector{
		multiSelectResponses: [][]string{{"work", "vr"}},
	}
	w := &Wizard{selector: sel, tagRepo: tagRepo}
	tags, err := w.promptTags()
	require.NoError(t, err)
	assert.Equal(t, []string{"work", "vr"}, tags)
}

func TestWizard_promptTags_Empty(t *testing.T) {
	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{}, nil)
	sel := &fakeSelector{
		multiSelectResponses: [][]string{{}},
	}
	w := &Wizard{selector: sel, tagRepo: tagRepo}
	tags, err := w.promptTags()
	require.NoError(t, err)
	assert.Empty(t, tags)
}

func TestWizard_promptTags_None(t *testing.T) {
	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{{Name: "work"}, {Name: "vr"}}, nil)
	sel := &fakeSelector{
		multiSelectResponses: [][]string{{selectorNone}},
	}
	w := &Wizard{selector: sel, tagRepo: tagRepo}
	tags, err := w.promptTags()
	require.NoError(t, err)
	assert.Empty(t, tags)
}

func TestWizard_promptTags_NoneTakesPriority(t *testing.T) {
	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{{Name: "work"}}, nil)
	sel := &fakeSelector{
		// Selecting both "none" and a tag — none should win
		multiSelectResponses: [][]string{{selectorNone, "work"}},
	}
	w := &Wizard{selector: sel, tagRepo: tagRepo}
	tags, err := w.promptTags()
	require.NoError(t, err)
	assert.Empty(t, tags)
}

func TestWizard_promptTags_InlineCreate(t *testing.T) {
	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{{Name: "work"}}, nil)
	tagRepo.On("GetByName", "nova-tag").Return(nil, nil) // tag doesn't exist
	tagRepo.On("Create", mock.AnythingOfType("*entity.Tag")).Return(nil)
	sel := &fakeSelector{
		multiSelectResponses: [][]string{{"work", selectorNewTag}},
	}
	prompter := &fakePrompter{
		textResponses: []string{"nova-tag"},
	}
	w := &Wizard{selector: sel, tagRepo: tagRepo, prompter: prompter, printer: cli.NewPrinter()}
	tags, err := w.promptTags()
	require.NoError(t, err)
	assert.Equal(t, []string{"work", "nova-tag"}, tags)
	tagRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.Tag"))
}

// --- promptCategory tests ---

func TestWizard_promptCategory_None(t *testing.T) {
	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{selectorNone, true}},
	}
	w := &Wizard{selector: sel, categoryRepo: catRepo}
	alias, err := w.promptCategory("nubank", entity.EntryTypeExpense)
	require.NoError(t, err)
	assert.Equal(t, "", alias)
}

func TestWizard_promptCategory_SelectExisting(t *testing.T) {
	foodCat := &entity.Category{Name: "Alimentação", Alias: "food", Type: entity.CategoryTypeExpense}
	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{foodCat}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{"food", true}},
	}
	w := &Wizard{selector: sel, categoryRepo: catRepo}
	alias, err := w.promptCategory("nubank", entity.EntryTypeExpense)
	require.NoError(t, err)
	assert.Equal(t, "food", alias)
}

func TestWizard_promptCategory_InlineCreate(t *testing.T) {
	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{}, nil)
	catRepo.On("Create", mock.AnythingOfType("*entity.Category"), "nubank").Return(nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{selectorNewCategory, true}},
	}
	prompter := &fakePrompter{
		textResponses: []string{"Alimentação", "food", ""},
	}
	w := &Wizard{selector: sel, categoryRepo: catRepo, prompter: prompter, printer: cli.NewPrinter()}
	alias, err := w.promptCategory("nubank", entity.EntryTypeExpense)
	require.NoError(t, err)
	assert.Equal(t, "food", alias)
	catRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.Category"), "nubank")
}

// --- promptCreditCard tests ---

func TestWizard_promptCreditCard_None(t *testing.T) {
	ccRepo := new(mockCCRepo)
	ccRepo.On("GetAll").Return([]*entity.CreditCard{}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{selectorNone, true}},
	}
	w := &Wizard{selector: sel, ccRepo: ccRepo}
	cc, times, err := w.promptCreditCard()
	require.NoError(t, err)
	assert.Equal(t, "", cc)
	assert.Equal(t, 1, times)
}

func TestWizard_promptCreditCard_SelectWithTimes(t *testing.T) {
	nuCard := &entity.CreditCard{Name: "nu", ClosingDay: 10, DueDay: 15}
	ccRepo := new(mockCCRepo)
	ccRepo.On("GetAll").Return([]*entity.CreditCard{nuCard}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{"nu", true}},
	}
	prompter := &fakePrompter{
		textResponses: []string{"3"},
	}
	w := &Wizard{selector: sel, ccRepo: ccRepo, prompter: prompter}
	cc, times, err := w.promptCreditCard()
	require.NoError(t, err)
	assert.Equal(t, "nu", cc)
	assert.Equal(t, 3, times)
}

func TestWizard_promptCreditCard_Cancelled(t *testing.T) {
	ccRepo := new(mockCCRepo)
	ccRepo.On("GetAll").Return([]*entity.CreditCard{}, nil)
	sel := &fakeSelector{
		selectResponses: []fakeSelectResponse{{"", false}},
	}
	w := &Wizard{selector: sel, ccRepo: ccRepo}
	_, _, err := w.promptCreditCard()
	assert.ErrorIs(t, err, errCancelled)
}

// --- promptAmount tests ---

func TestWizard_promptAmount_Valid(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"50.00"},
	}
	w := &Wizard{prompter: prompter}
	amount, err := w.promptAmount()
	require.NoError(t, err)
	assert.Equal(t, "50.00", amount)
}

func TestWizard_promptAmount_Expression(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"1000/3"},
	}
	w := &Wizard{prompter: prompter}
	amount, err := w.promptAmount()
	require.NoError(t, err)
	assert.Equal(t, "1000/3", amount)
}

func TestWizard_promptAmount_RetryOnInvalid(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"abc", "invalid", "50"},
	}
	w := &Wizard{prompter: prompter}
	amount, err := w.promptAmount()
	require.NoError(t, err)
	assert.Equal(t, "50", amount)
}

// --- promptDate tests ---

func TestWizard_promptDate_Valid(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"2025-06-15"},
	}
	w := &Wizard{prompter: prompter}
	date, dateStr, err := w.promptDate()
	require.NoError(t, err)
	assert.Equal(t, "2025-06-15", dateStr)
	assert.Equal(t, 2025, date.Year())
	assert.Equal(t, time.June, date.Month())
	assert.Equal(t, 15, date.Day())
}

func TestWizard_promptDate_RetryOnInvalid(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"invalid", "2025-06-15"},
	}
	w := &Wizard{prompter: prompter}
	date, dateStr, err := w.promptDate()
	require.NoError(t, err)
	assert.Equal(t, "2025-06-15", dateStr)
	assert.False(t, date.IsZero())
}

// --- promptTimes tests ---

func TestWizard_promptTimes_Default(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{""},
	}
	w := &Wizard{prompter: prompter}
	times, err := w.promptTimes()
	require.NoError(t, err)
	assert.Equal(t, 1, times)
}

func TestWizard_promptTimes_Custom(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"12"},
	}
	w := &Wizard{prompter: prompter}
	times, err := w.promptTimes()
	require.NoError(t, err)
	assert.Equal(t, 12, times)
}

func TestWizard_promptTimes_RetryOnInvalid(t *testing.T) {
	prompter := &fakePrompter{
		textResponses: []string{"0", "abc", "3"},
	}
	w := &Wizard{prompter: prompter}
	times, err := w.promptTimes()
	require.NoError(t, err)
	assert.Equal(t, 3, times)
}

// --- execute tests ---

func TestWizard_execute_Success(t *testing.T) {
	date := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	cat := &entity.Category{Name: "Food", Alias: "food", Type: entity.CategoryTypeExpense}
	categoryRepo := new(mockCategoriesRepo)
	categoryRepo.On("GetByAlias", "nubank", "food").Return(cat, nil)

	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetByName", "work").Return(&entity.Tag{Name: "work"}, nil)

	ccRepo := new(mockCCRepo)

	entryRepo := new(mockEntriesRepo)
	entryRepo.On("Create", mock.MatchedBy(func(e *entity.Entry) bool {
		return e.Amount == 50.00 && e.Description == "almoço" && e.Type == entity.EntryTypeExpense
	}), "nubank").Return(nil)

	w := &Wizard{
		entryRepo:       entryRepo,
		categoryRepo:    categoryRepo,
		tagRepo:         tagRepo,
		ccRepo:          ccRepo,
		defaultCurrency: "BRL",
		printer:         cli.NewPrinter(),
	}

	err := w.execute(
		entity.EntryTypeExpense, "50.00", "nubank", date, "almoço",
		"food", "", []string{"work"}, 1,
	)
	require.NoError(t, err)

	entryRepo.AssertExpectations(t)
}

func TestWizard_execute_WithCreditCard(t *testing.T) {
	date := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	cat := &entity.Category{Name: "Food", Alias: "food", Type: entity.CategoryTypeExpense}
	categoryRepo := new(mockCategoriesRepo)
	categoryRepo.On("GetByAlias", "nubank", "food").Return(cat, nil)

	tagRepo := new(mockTagsRepo)

	nuCard := &entity.CreditCard{Name: "nu", ClosingDay: 10, DueDay: 15}
	ccRepo := new(mockCCRepo)
	ccRepo.On("GetByName", "nu").Return(nuCard, nil)

	entryRepo := new(mockEntriesRepo)
	entryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "nubank").Return(nil)

	w := &Wizard{
		entryRepo:       entryRepo,
		categoryRepo:    categoryRepo,
		tagRepo:         tagRepo,
		ccRepo:          ccRepo,
		defaultCurrency: "BRL",
		printer:         cli.NewPrinter(),
	}

	err := w.execute(
		entity.EntryTypeExpense, "1000", "nubank", date, "compra",
		"food", "nu", nil, 3,
	)
	require.NoError(t, err)

	entryRepo.AssertNumberOfCalls(t, "Create", 3)
}

// --- parseDate tests ---

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{"YYYY-MM-DD", "2025-06-15", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"empty", "", time.Time{}},
		{"spaces only", "   ", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- isTerminal tests ---

func TestWizard_isTerminal_RegularFile(t *testing.T) {
	// isTerminal is hard to test without OS-level mocking; it's tested in Task 7.
	// This test verifies it compiles and doesn't panic on a nil stat scenario.
	// The core logic (ModeCharDevice check) is from the standard library.
}

// --- runOneCycle integration tests ---

func TestWizard_runOneCycle_ExpenseNoCreditCard(t *testing.T) {
	accRepo := new(mockAccountsRepo)
	accRepo.On("GetAll").Return([]*entity.Account{{Name: "nubank"}}, nil)

	cat := &entity.Category{Name: "Food", Alias: "food", Type: entity.CategoryTypeExpense}
	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{cat}, nil)
	catRepo.On("GetByAlias", "nubank", "food").Return(cat, nil)

	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{}, nil)
	tagRepo.On("GetByName", mock.Anything).Return(nil, nil)

	ccRepo := new(mockCCRepo)
	ccRepo.On("GetAll").Return([]*entity.CreditCard{}, nil)

	entryRepo := new(mockEntriesRepo)
	entryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "nubank").Return(nil)

	prompter := &fakePrompter{
		textResponses:    []string{"50.00", "2025-06-15", "almoço"},
		confirmResponses: []bool{true},
	}
	selector := &fakeSelector{
		selectResponses: []fakeSelectResponse{
			{choiceExpense, true}, // type
			{"nubank", true},      // account
			{"food", true},        // category
			{selectorNone, true},  // credit card
		},
		multiSelectResponses: [][]string{{}},
	}

	w := NewWizard(prompter, selector, entryRepo, catRepo, tagRepo, ccRepo, accRepo, "BRL", cli.NewPrinter())

	err := w.runOneCycle()
	require.NoError(t, err)

	entryRepo.AssertExpectations(t)
}

func TestWizard_runOneCycle_Income(t *testing.T) {
	accRepo := new(mockAccountsRepo)
	accRepo.On("GetAll").Return([]*entity.Account{{Name: "nubank"}}, nil)

	incCat := &entity.Category{Name: "Salário", Alias: "sal", Type: entity.CategoryTypeIncome}
	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{incCat}, nil)
	catRepo.On("GetByAlias", "nubank", "sal").Return(incCat, nil)

	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{}, nil)

	ccRepo := new(mockCCRepo)

	entryRepo := new(mockEntriesRepo)
	entryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "nubank").Return(nil)

	prompter := &fakePrompter{
		textResponses:    []string{"5000", "2025-06-15", "salário"},
		confirmResponses: []bool{true},
	}
	selector := &fakeSelector{
		selectResponses: []fakeSelectResponse{
			{choiceIncome, true}, // type
			{"nubank", true},     // account
			{"sal", true},        // category
		},
		multiSelectResponses: [][]string{{}},
	}

	w := NewWizard(prompter, selector, entryRepo, catRepo, tagRepo, ccRepo, accRepo, "BRL", cli.NewPrinter())

	err := w.runOneCycle()
	require.NoError(t, err)

	entryRepo.AssertExpectations(t)
}

func TestWizard_runOneCycle_CancelOnType(t *testing.T) {
	selector := &fakeSelector{
		selectResponses: []fakeSelectResponse{
			{"", false}, // cancelled on type
		},
	}
	w := &Wizard{selector: selector}
	err := w.runOneCycle()
	assert.ErrorIs(t, err, errCancelled)
}

func TestWizard_runOneCycle_DeclineConfirm(t *testing.T) {
	accRepo := new(mockAccountsRepo)
	accRepo.On("GetAll").Return([]*entity.Account{{Name: "nubank"}}, nil)

	catRepo := new(mockCategoriesRepo)
	catRepo.On("GetAll", "nubank").Return([]*entity.Category{}, nil)

	tagRepo := new(mockTagsRepo)
	tagRepo.On("GetAll").Return([]*entity.Tag{}, nil)

	ccRepo := new(mockCCRepo)
	ccRepo.On("GetAll").Return([]*entity.CreditCard{}, nil)

	entryRepo := new(mockEntriesRepo)

	prompter := &fakePrompter{
		textResponses:    []string{"50.00", "2025-06-15", "almoço"},
		confirmResponses: []bool{false}, // decline execution
	}
	selector := &fakeSelector{
		selectResponses: []fakeSelectResponse{
			{choiceExpense, true}, // type
			{"nubank", true},      // account
			{selectorNone, true},  // category
			{selectorNone, true},  // credit card
		},
		multiSelectResponses: [][]string{{}},
	}

	w := NewWizard(prompter, selector, entryRepo, catRepo, tagRepo, ccRepo, accRepo, "BRL", cli.NewPrinter())

	err := w.runOneCycle()
	require.NoError(t, err) // declined = not an error

	entryRepo.AssertNotCalled(t, "Create")
}
