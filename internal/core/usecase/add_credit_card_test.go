package usecase

import (
	"testing"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCreditCardsRepository struct {
	mock.Mock
}

func (m *MockCreditCardsRepository) Create(cc *entity.CreditCard) error {
	args := m.Called(cc)
	return args.Error(0)
}

func (m *MockCreditCardsRepository) GetByName(name string) (*entity.CreditCard, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CreditCard), args.Error(1)
}

func (m *MockCreditCardsRepository) GetAll() ([]*entity.CreditCard, error) {
	args := m.Called()
	return args.Get(0).([]*entity.CreditCard), args.Error(1)
}

func TestAddCreditCard_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockCreditCardsRepository)
		uc := NewAddCreditCard(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*entity.CreditCard")).Return(nil)

		result, err := uc.Execute(AddCreditCardInput{
			Name:       "Nubank",
			ClosingDay: 9,
			DueDay:     16,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Nubank", result.CreditCard.Name)
		assert.Equal(t, 9, result.CreditCard.ClosingDay)
		assert.Equal(t, 16, result.CreditCard.DueDay)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid closing day", func(t *testing.T) {
		mockRepo := new(MockCreditCardsRepository)
		uc := NewAddCreditCard(mockRepo)

		result, err := uc.Execute(AddCreditCardInput{
			Name:       "Nubank",
			ClosingDay: 0,
			DueDay:     16,
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, entity.ErrInvalidClosingDay)
	})
}
