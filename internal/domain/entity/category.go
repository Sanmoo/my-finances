package entity

import "errors"

var (
	ErrInvalidCategoryType = errors.New("category type must be 'inc' or 'exp'")
	ErrEmptyCategoryName   = errors.New("category name cannot be empty")
	ErrEmptyCategoryAlias  = errors.New("category alias cannot be empty")
	ErrCategoryNotFound    = errors.New("category not found")
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "inc"
	CategoryTypeExpense CategoryType = "exp"
)

type Category struct {
	Name  string
	Alias string
	Emoji *string
	Type  CategoryType
}

func NewCategory(name, alias string, catType CategoryType, opts ...CategoryOption) (*Category, error) {
	name = TrimLower(name)
	if name == "" {
		return nil, ErrEmptyCategoryName
	}

	alias = TrimLower(alias)
	if alias == "" {
		return nil, ErrEmptyCategoryAlias
	}

	if catType != CategoryTypeIncome && catType != CategoryTypeExpense {
		return nil, ErrInvalidCategoryType
	}

	cat := &Category{
		Name:  name,
		Alias: alias,
		Type:  catType,
	}

	for _, opt := range opts {
		opt(cat)
	}

	return cat, nil
}

type CategoryOption func(*Category)

func WithEmoji(emoji string) CategoryOption {
	return func(c *Category) {
		if emoji != "" {
			c.Emoji = &emoji
		}
	}
}

func (c *Category) IsIncome() bool {
	return c.Type == CategoryTypeIncome
}

func (c *Category) IsExpense() bool {
	return c.Type == CategoryTypeExpense
}
