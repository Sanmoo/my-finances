package entity

import "errors"

var (
	ErrInvalidCategoryType = errors.New("category type must be 'inc' or 'exp'")
	ErrEmptyCategoryName   = errors.New("category name cannot be empty")
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "inc"
	CategoryTypeExpense CategoryType = "exp"
)

type Category struct {
	ID    int64
	Name  string
	Alias *string
	Emoji *string
	Type  CategoryType
}

func NewCategory(name string, catType CategoryType, opts ...CategoryOption) (*Category, error) {
	name = trimLower(name)
	if name == "" {
		return nil, ErrEmptyCategoryName
	}
	if catType != CategoryTypeIncome && catType != CategoryTypeExpense {
		return nil, ErrInvalidCategoryType
	}

	cat := &Category{
		Name: name,
		Type: catType,
	}

	for _, opt := range opts {
		opt(cat)
	}

	return cat, nil
}

type CategoryOption func(*Category)

func WithAlias(alias string) CategoryOption {
	return func(c *Category) {
		alias = trimLower(alias)
		if alias != "" {
			c.Alias = &alias
		}
	}
}

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
