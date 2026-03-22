package usecase

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AddCategoryInput struct {
	AccountName string
	Name        string
	Type        entity.CategoryType
	Alias       string
	Emoji       string
}

type AddCategoryOutput struct {
	Category *entity.Category
}

type AddCategory struct {
	repo port.CategoriesRepository
}

func NewAddCategory(repo port.CategoriesRepository) *AddCategory {
	return &AddCategory{repo: repo}
}

func (uc *AddCategory) Execute(input AddCategoryInput) (*AddCategoryOutput, error) {
	var opts []entity.CategoryOption
	if input.Emoji != "" {
		opts = append(opts, entity.WithEmoji(input.Emoji))
	}

	category, err := entity.NewCategory(input.Name, input.Alias, input.Type, opts...)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(category, input.AccountName); err != nil {
		return nil, err
	}

	return &AddCategoryOutput{Category: category}, nil
}
