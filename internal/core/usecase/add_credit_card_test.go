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

func (m *MockCreditCardsRepository) Create(cc *entity.CreditCard) (int64, error) {
	args := m.Called(cc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCreditCardsRepository) GetByID(id int64) (*entity.CreditCard, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CreditCard), args.Error(1)
}

func (m *MockCreditCardsRepository) GetByNamespaceID(namespaceID int64) ([]*entity.CreditCard, error) {
	args := m.Called(namespaceID)
	return args.Get(0).([]*entity.CreditCard), args.Error(1)
}

func (m *MockCreditCardsRepository) Update(cc *entity.CreditCard) error {
	args := m.Called(cc)
	return args.Error(0)
}

func (m *MockCreditCardsRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAddCreditCard_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockCreditCardsRepository)
		uc := NewAddCreditCard(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*entity.CreditCard")).Return(int64(1), nil)

		result, err := uc.Execute(AddCreditCardInput{
			NamespaceID: 1,
			Name:        "Nubank",
			ClosingDay:  9,
			DueDay:      16,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.CreditCard.ID)
		assert.Equal(t, "Nubank", result.CreditCard.Name)
		assert.Equal(t, 9, result.CreditCard.ClosingDay)
		assert.Equal(t, 16, result.CreditCard.DueDay)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid closing day", func(t *testing.T) {
		mockRepo := new(MockCreditCardsRepository)
		uc := NewAddCreditCard(mockRepo)

		result, err := uc.Execute(AddCreditCardInput{
			NamespaceID: 1,
			Name:        "Nubank",
			ClosingDay:  0,
			DueDay:      16,
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, entity.ErrInvalidClosingDay)
	})
}
