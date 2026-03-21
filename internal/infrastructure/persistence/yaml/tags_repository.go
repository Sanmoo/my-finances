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

type Tag struct {
	ID   int64  `yaml:"id"`
	Name string `yaml:"name"`
}

type TagsData struct {
	Tags []Tag `yaml:"tags"`
}

func NewTagsRepository(basePath string) *TagsRepository {
	return &TagsRepository{basePath: basePath}
}

func (r *TagsRepository) filePath() string {
	return filepath.Join(r.basePath, "tags.yaml")
}

func (r *TagsRepository) metaPath() string {
	return filepath.Join(r.basePath, "_meta.yaml")
}

func (r *TagsRepository) Create(tag *entity.Tag) (int64, error) {
	if err := EnsureMetaFile(r.metaPath()); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(), "tags")
	if err != nil {
		return 0, err
	}

	tag.ID = nextID

	data := &TagsData{}
	if _, err := os.Stat(r.filePath()); err == nil {
		readData, err := Read[TagsData](r.filePath())
		if err != nil {
			return 0, err
		}
		data = readData
	}

	yamlTag := Tag{
		ID:   tag.ID,
		Name: tag.Name,
	}
	data.Tags = append(data.Tags, yamlTag)

	if err := Write(r.filePath(), data); err != nil {
		return 0, err
	}

	return tag.ID, nil
}

func (r *TagsRepository) GetByID(id int64) (*entity.Tag, error) {
	data, err := Read[TagsData](r.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, t := range data.Tags {
		if t.ID == id {
			return &entity.Tag{
				ID:   t.ID,
				Name: t.Name,
			}, nil
		}
	}

	return nil, nil
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
				ID:   t.ID,
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
			ID:   t.ID,
			Name: t.Name,
		})
	}

	return tags, nil
}

func (r *TagsRepository) EnsureInitialized() error {
	return nil
}

func (r *TagsRepository) Port() port.TagsRepository {
	return r
}
