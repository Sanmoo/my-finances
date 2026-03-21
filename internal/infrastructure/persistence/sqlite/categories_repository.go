package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CategoriesRepository struct {
	db *DB
}

func NewCategoriesRepository(db *DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) Create(cat *entity.Category) (int64, error) {
	query := `INSERT INTO categories (namespace_id, name, alias, emoji, type) VALUES (?, ?, ?, ?, ?)`

	var alias, emoji interface{}
	if cat.Alias != nil {
		alias = *cat.Alias
	}
	if cat.Emoji != nil {
		emoji = *cat.Emoji
	}

	result, err := r.db.Exec(query, cat.NamespaceID, cat.Name, alias, emoji, cat.Type)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *CategoriesRepository) GetByID(id int64) (*entity.Category, error) {
	query := `SELECT id, namespace_id, name, alias, emoji, type FROM categories WHERE id = ?`

	var cat entity.Category
	var alias, emoji sql.NullString

	err := r.db.QueryRow(query, id).Scan(&cat.ID, &cat.NamespaceID, &cat.Name, &alias, &emoji, &cat.Type)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if alias.Valid {
		cat.Alias = &alias.String
	}
	if emoji.Valid {
		cat.Emoji = &emoji.String
	}

	return &cat, nil
}

func (r *CategoriesRepository) GetByNamespaceID(namespaceID int64) ([]*entity.Category, error) {
	query := `SELECT id, namespace_id, name, alias, emoji, type FROM categories WHERE namespace_id = ? ORDER BY name`

	rows, err := r.db.Query(query, namespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		var cat entity.Category
		var alias, emoji sql.NullString

		if err := rows.Scan(&cat.ID, &cat.NamespaceID, &cat.Name, &alias, &emoji, &cat.Type); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		if alias.Valid {
			cat.Alias = &alias.String
		}
		if emoji.Valid {
			cat.Emoji = &emoji.String
		}

		categories = append(categories, &cat)
	}

	return categories, nil
}

func (r *CategoriesRepository) GetByNameOrAlias(namespaceID int64, nameOrAlias string) (*entity.Category, error) {
	query := `SELECT id, namespace_id, name, alias, emoji, type FROM categories WHERE namespace_id = ? AND (name = ? OR alias = ?)`

	var cat entity.Category
	var alias, emoji sql.NullString

	err := r.db.QueryRow(query, namespaceID, nameOrAlias, nameOrAlias).Scan(&cat.ID, &cat.NamespaceID, &cat.Name, &alias, &emoji, &cat.Type)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if alias.Valid {
		cat.Alias = &alias.String
	}
	if emoji.Valid {
		cat.Emoji = &emoji.String
	}

	return &cat, nil
}

func (r *CategoriesRepository) Update(cat *entity.Category) error {
	query := `UPDATE categories SET namespace_id = ?, name = ?, alias = ?, emoji = ?, type = ? WHERE id = ?`

	var alias, emoji interface{}
	if cat.Alias != nil {
		alias = *cat.Alias
	}
	if cat.Emoji != nil {
		emoji = *cat.Emoji
	}

	_, err := r.db.Exec(query, cat.NamespaceID, cat.Name, alias, emoji, cat.Type, cat.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (r *CategoriesRepository) Delete(id int64) error {
	query := `DELETE FROM categories WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
