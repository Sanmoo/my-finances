package yaml

import (
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type TagsRepository struct {
	basePath string
}

func NewTagsRepository(basePath string) *TagsRepository {
	return &TagsRepository{basePath: basePath}
}

func (r *TagsRepository) filePath() string {
	return filepath.Join(r.basePath, "tags.yaml")
}

func (r *TagsRepository) Create(tag *entity.Tag) error {
	data, err := Read[TagsData](r.filePath())
	if err != nil {
		return err
	}

	yamlTag := Tag{}
	yamlTag.FromEntity(tag)
	data.Tags = append(data.Tags, yamlTag)

	return Write(r.filePath(), data)
}

func (r *TagsRepository) GetByName(name string) (*entity.Tag, error) {
	name = entity.TrimLower(name)

	data, err := Read[TagsData](r.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, t := range data.Tags {
		if t.Name == name {
			return &entity.Tag{
				Name: t.Name,
			}, nil
		}
	}

	return nil, nil
}

func (r *TagsRepository) GetAll() ([]*entity.Tag, error) {
	data, err := Read[TagsData](r.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*entity.Tag{}, nil
		}
		return nil, err
	}

	tags := make([]*entity.Tag, 0, len(data.Tags))
	for _, t := range data.Tags {
		tags = append(tags, &entity.Tag{
			Name: t.Name,
		})
	}

	return tags, nil
}

func (r *TagsRepository) EnsureInitialized() error {
	path := r.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Write(path, TagsData{Tags: []Tag{}})
	}
	return nil
}

func (r *TagsRepository) Port() port.TagsRepository {
	return r
}
