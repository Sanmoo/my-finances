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

func (m *MockCategoriesRepository) Create(cat *entity.Category) (int64, error) {
	args := m.Called(cat)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCategoriesRepository) GetByID(id int64) (*entity.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoriesRepository) GetByNamespaceID(namespaceID int64) ([]*entity.Category, error) {
	args := m.Called(namespaceID)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func (m *MockCategoriesRepository) GetByNameOrAlias(namespaceID int64, nameOrAlias string) (*entity.Category, error) {
	args := m.Called(namespaceID, nameOrAlias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoriesRepository) Update(cat *entity.Category) error {
	args := m.Called(cat)
	return args.Error(0)
}

func (m *MockCategoriesRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAddCategory_Execute(t *testing.T) {
	t.Run("success with all options", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		uc := NewAddCategory(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*entity.Category")).Return(int64(1), nil)

		result, err := uc.Execute(AddCategoryInput{
			NamespaceID: 1,
			Name:        "Food",
			Type:        entity.CategoryTypeExpense,
			Alias:       "food",
			Emoji:       "🍔",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.Category.ID)
		assert.Equal(t, "food", result.Category.Name)
		assert.NotNil(t, result.Category.Alias)
		assert.Equal(t, "food", *result.Category.Alias)
		assert.NotNil(t, result.Category.Emoji)
		assert.Equal(t, "🍔", *result.Category.Emoji)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid type", func(t *testing.T) {
		mockRepo := new(MockCategoriesRepository)
		uc := NewAddCategory(mockRepo)

		result, err := uc.Execute(AddCategoryInput{
			NamespaceID: 1,
			Name:        "Test",
			Type:        "invalid",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
