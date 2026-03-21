package entity

import "errors"

var (
	ErrEmptyTagName = errors.New("tag name cannot be empty")
)

type Tag struct {
	ID   int64
	Name string
}

func NewTag(name string) (*Tag, error) {
	name = TrimLower(name)
	if name == "" {
		return nil, ErrEmptyTagName
	}

	return &Tag{
		Name: name,
	}, nil
}
