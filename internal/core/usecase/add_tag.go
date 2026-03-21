package usecase

import (
	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AddTagInput struct {
	Name string
}

type AddTagOutput struct {
	Tag *entity.Tag
}

type AddTag struct {
	repo port.TagsRepository
}

func NewAddTag(repo port.TagsRepository) *AddTag {
	return &AddTag{repo: repo}
}

func (uc *AddTag) Execute(input AddTagInput) (*AddTagOutput, error) {
	tag, err := entity.NewTag(input.Name)
	if err != nil {
		return nil, err
	}

	id, err := uc.repo.Create(tag)
	if err != nil {
		return nil, err
	}

	tag.ID = id
	return &AddTagOutput{Tag: tag}, nil
}
