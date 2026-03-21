package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name    string
		catName string
		catType CategoryType
		opts    []CategoryOption
		wantErr error
	}{
		{
			name:    "valid income category",
			catName: "Salary",
			catType: CategoryTypeIncome,
			opts:    []CategoryOption{},
			wantErr: nil,
		},
		{
			name:    "valid expense category",
			catName: "Food",
			catType: CategoryTypeExpense,
			opts:    []CategoryOption{},
			wantErr: nil,
		},
		{
			name:    "invalid category type",
			catName: "Test",
			catType: "invalid",
			wantErr: ErrInvalidCategoryType,
		},
		{
			name:    "empty name",
			catName: "",
			catType: CategoryTypeExpense,
			wantErr: ErrEmptyCategoryName,
		},
		{
			name:    "with alias",
			catName: "Transport",
			catType: CategoryTypeExpense,
			opts:    []CategoryOption{WithAlias("transport")},
			wantErr: nil,
		},
		{
			name:    "with emoji",
			catName: "Food",
			catType: CategoryTypeExpense,
			opts:    []CategoryOption{WithEmoji("🍔")},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := NewCategory(tt.catName, tt.catType, tt.opts...)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cat)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cat)
				assert.Equal(t, tt.catType, cat.Type)
			}
		})
	}
}

func TestCategory_IsIncome(t *testing.T) {
	cat := &Category{Type: CategoryTypeIncome}
	assert.True(t, cat.IsIncome())
	assert.False(t, cat.IsExpense())

	cat.Type = CategoryTypeExpense
	assert.False(t, cat.IsIncome())
	assert.True(t, cat.IsExpense())
}
