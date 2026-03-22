package usecase

import (
	"testing"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountsRepository struct {
	mock.Mock
}

func (m *MockAccountsRepository) Create(acc *entity.Account) error {
	args := m.Called(acc)
	return args.Error(0)
}

func (m *MockAccountsRepository) GetByName(name string) (*entity.Account, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *MockAccountsRepository) GetAll() ([]*entity.Account, error) {
	args := m.Called()
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func TestAddAccount_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockAccountsRepository)
		uc := NewAddAccount(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*entity.Account")).Return(nil)

		result, err := uc.Execute(AddAccountInput{
			Name: "main",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "main", result.Account.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid name", func(t *testing.T) {
		mockRepo := new(MockAccountsRepository)
		uc := NewAddAccount(mockRepo)

		result, err := uc.Execute(AddAccountInput{
			Name: "",
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, entity.ErrEmptyAccountName)
	})
}
