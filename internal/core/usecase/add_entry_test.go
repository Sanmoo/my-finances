package usecase

import (
	"testing"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEntriesRepository struct {
	mock.Mock
}

func (m *MockEntriesRepository) Create(entry *entity.Entry, accountName string) error {
	args := m.Called(entry, accountName)
	return args.Error(0)
}

func (m *MockEntriesRepository) GetAll(filters *port.EntryFilters) ([]*entity.Entry, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Entry), args.Error(1)
}

type MockTagsRepository struct {
	mock.Mock
}

func (m *MockTagsRepository) Create(tag *entity.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

func (m *MockTagsRepository) GetByName(name string) (*entity.Tag, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Tag), args.Error(1)
}

func (m *MockTagsRepository) GetAll() ([]*entity.Tag, error) {
	args := m.Called()
	return args.Get(0).([]*entity.Tag), args.Error(1)
}

func TestAddEntry_Execute(t *testing.T) {
	t.Run("success with simple amount", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockTagRepo := new(MockTagsRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockTagRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "test-account").Return(nil)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			Type:        entity.EntryTypeExpense,
			Amount:      "50.00",
			Currency:    "BRL",
			Description: "Test expense",
			Date:        date,
			AccountName: "test-account",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 1)
		assert.Equal(t, 50.00, result.Entries[0].Amount)

		mockEntryRepo.AssertExpectations(t)
	})

	t.Run("success with math expression", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockTagRepo := new(MockTagsRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockTagRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "test-account").Return(nil)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			Type:        entity.EntryTypeIncome,
			Amount:      "1000/2",
			Currency:    "BRL",
			Description: "Test income",
			Date:        date,
			AccountName: "test-account",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 1)
		assert.Equal(t, 500.00, result.Entries[0].Amount)
	})

	t.Run("success with multiple installments", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockTagRepo := new(MockTagsRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockTagRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "test-account").Return(nil).Once()
		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry"), "test-account").Return(nil).Once()

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			Type:        entity.EntryTypeExpense,
			Amount:      "100",
			Currency:    "BRL",
			Description: "Test installments",
			Date:        date,
			Times:       2,
			AccountName: "test-account",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 2)

		mockEntryRepo.AssertExpectations(t)
	})

	t.Run("invalid amount expression", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockTagRepo := new(MockTagsRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockTagRepo, mockCCRepo)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			Type:        entity.EntryTypeExpense,
			Amount:      "invalid",
			Currency:    "BRL",
			Description: "Test invalid",
			Date:        date,
			AccountName: "test-account",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
