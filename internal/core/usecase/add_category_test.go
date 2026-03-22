package usecase

import (
	"testing"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoriesRepository struct {
	mock.Mock
}

func (m *MockCategoriesRepository) Create(cat *entity.Category, accountName string) error {
	args := m.Called(cat, accountName)
	return args.Error(0)
}

func (m *MockCategoriesRepository) GetByAlias(accountName string, alias string) (*entity.Category, error) {
	args := m.Called(accountName, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoriesRepository) GetAll(accountName string) ([]*entity.Category, error) {
	args := m.Called(accountName)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func TestAddCategory_Execute(t *testing.T) {
	t.Run("success with all options", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		uc := NewAddCategory(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*entity.Category"), "test-account").Return(nil)

		result, err := uc.Execute(AddCategoryInput{
			AccountName: "test-account",
			Name:        "Food",
			Type:        entity.CategoryTypeExpense,
			Alias:       "food",
			Emoji:       "🍔",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "food", result.Category.Name)
		assert.Equal(t, "food", result.Category.Alias)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid type", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		uc := NewAddCategory(mockRepo)

		result, err := uc.Execute(AddCategoryInput{
			AccountName: "test-account",
			Name:        "Test",
			Alias:       "test",
			Type:        "invalid",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
