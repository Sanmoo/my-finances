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

func (m *MockEntriesRepository) Create(entry *entity.Entry) (int64, error) {
	args := m.Called(entry)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEntriesRepository) GetByID(id int64) (*entity.Entry, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Entry), args.Error(1)
}

func (m *MockEntriesRepository) GetByNamespaceID(namespaceID int64, filters *port.EntryFilters) ([]*entity.Entry, error) {
	args := m.Called(namespaceID, filters)
	return args.Get(0).([]*entity.Entry), args.Error(1)
}

func (m *MockEntriesRepository) Update(entry *entity.Entry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockEntriesRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEntriesRepository) AddTag(entryID int64, tag string) error {
	args := m.Called(entryID, tag)
	return args.Error(0)
}

func (m *MockEntriesRepository) RemoveTag(entryID int64, tag string) error {
	args := m.Called(entryID, tag)
	return args.Error(0)
}

func TestAddEntry_Execute(t *testing.T) {
	t.Run("success with simple amount", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry")).Return(int64(1), nil)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			NamespaceID: 1,
			Type:        entity.EntryTypeExpense,
			Amount:      "50.00",
			Currency:    "BRL",
			Date:        date,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 1)
		assert.Equal(t, int64(1), result.Entries[0].ID)
		assert.Equal(t, 50.00, result.Entries[0].Amount)

		mockEntryRepo.AssertExpectations(t)
	})

	t.Run("success with math expression", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry")).Return(int64(1), nil)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			NamespaceID: 1,
			Type:        entity.EntryTypeIncome,
			Amount:      "1000/2",
			Currency:    "BRL",
			Date:        date,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 1)
		assert.Equal(t, 500.00, result.Entries[0].Amount)
	})

	t.Run("success with multiple installments", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockCCRepo)

		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry")).Return(int64(1), nil).Once()
		mockEntryRepo.On("Create", mock.AnythingOfType("*entity.Entry")).Return(int64(2), nil).Once()

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			NamespaceID: 1,
			Type:        entity.EntryTypeExpense,
			Amount:      "100",
			Currency:    "BRL",
			Date:        date,
			Times:       2,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Entries, 2)

		mockEntryRepo.AssertExpectations(t)
	})

	t.Run("invalid amount expression", func(t *testing.T) {
		mockEntryRepo := new(MockEntriesRepository)
		mockCategoryRepo := new(MockCategoriesRepository)
		mockCCRepo := new(MockCreditCardsRepository)
		uc := NewAddEntry(mockEntryRepo, mockCategoryRepo, mockCCRepo)

		date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		result, err := uc.Execute(AddEntryInput{
			NamespaceID: 1,
			Type:        entity.EntryTypeExpense,
			Amount:      "invalid",
			Currency:    "BRL",
			Date:        date,
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
