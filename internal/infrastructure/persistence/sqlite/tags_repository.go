package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type TagsRepository struct {
	db *DB
}

func NewTagsRepository(db *DB) *TagsRepository {
	return &TagsRepository{db: db}
}

func (r *TagsRepository) Create(tag *entity.Tag) (int64, error) {
	query := `INSERT INTO tags (name) VALUES (?)`

	result, err := r.db.Exec(query, tag.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to create tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *TagsRepository) GetByID(id int64) (*entity.Tag, error) {
	query := `SELECT id, name FROM tags WHERE id = ?`

	var tag entity.Tag
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &tag, nil
}

func (r *TagsRepository) GetByName(name string) (*entity.Tag, error) {
	name = entity.TrimLower(name)

	query := `SELECT id, name FROM tags WHERE name = ?`

	var tag entity.Tag
	err := r.db.QueryRow(query, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return &tag, nil
}

func (r *TagsRepository) GetAll() ([]*entity.Tag, error) {
	query := `SELECT id, name FROM tags ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []*entity.Tag
	for rows.Next() {
		var tag entity.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &tag)
	}

	return tags, nil
}
