package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type EntriesRepository struct {
	db *DB
}

func NewEntriesRepository(db *DB) *EntriesRepository {
	return &EntriesRepository{db: db}
}

func (r *EntriesRepository) Create(entry *entity.Entry) (int64, error) {
	query := `INSERT INTO entries (namespace_id, type, amount, currency, description, category_id, credit_card_id, realization_date, payment_date, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var categoryID, creditCardID interface{}
	if entry.CategoryID != nil {
		categoryID = *entry.CategoryID
	}
	if entry.CreditCardID != nil {
		creditCardID = *entry.CreditCardID
	}

	var paymentDate interface{}
	if entry.PaymentDate != nil {
		paymentDate = entry.PaymentDate.Format("2006-01-02")
	}

	result, err := r.db.Exec(query,
		entry.NamespaceID,
		entry.Type,
		entry.Amount,
		entry.Currency,
		entry.Description,
		categoryID,
		creditCardID,
		entry.RealizationDate.Format("2006-01-02"),
		paymentDate,
		entry.CreatedAt.Format("2006-01-02"),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	for _, tag := range entry.Tags {
		if err := r.AddTag(id, tag); err != nil {
			return 0, fmt.Errorf("failed to add tag: %w", err)
		}
	}

	return id, nil
}

func (r *EntriesRepository) GetByID(id int64) (*entity.Entry, error) {
	query := `SELECT id, namespace_id, type, amount, currency, description, category_id, credit_card_id, realization_date, payment_date, created_at 
			  FROM entries WHERE id = ?`

	var entry entity.Entry
	var description sql.NullString
	var categoryID, creditCardID sql.NullInt64
	var paymentDate sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.NamespaceID,
		&entry.Type,
		&entry.Amount,
		&entry.Currency,
		&description,
		&categoryID,
		&creditCardID,
		&entry.RealizationDate,
		&paymentDate,
		&entry.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if description.Valid {
		entry.Description = description.String
	}
	if categoryID.Valid {
		entry.CategoryID = &categoryID.Int64
	}
	if creditCardID.Valid {
		entry.CreditCardID = &creditCardID.Int64
	}
	if paymentDate.Valid {
		t, _ := time.Parse("2006-01-02", paymentDate.String)
		entry.PaymentDate = &t
	}

	tags, err := r.getTagsByEntryID(entry.ID)
	if err != nil {
		return nil, err
	}
	entry.Tags = tags

	return &entry, nil
}

func (r *EntriesRepository) GetByNamespaceID(namespaceID int64, filters *port.EntryFilters) ([]*entity.Entry, error) {
	query := `SELECT id, namespace_id, type, amount, currency, description, category_id, credit_card_id, realization_date, payment_date, created_at 
			  FROM entries WHERE namespace_id = ?`
	var args []interface{}
	args = append(args, namespaceID)

	if filters != nil {
		if filters.Type != nil {
			query += " AND type = ?"
			args = append(args, *filters.Type)
		}
		if filters.FromDate != nil {
			query += " AND realization_date >= ?"
			args = append(args, filters.FromDate.Format("2006-01-02"))
		}
		if filters.ToDate != nil {
			query += " AND realization_date <= ?"
			args = append(args, filters.ToDate.Format("2006-01-02"))
		}
		if filters.CreditCardID != nil {
			query += " AND credit_card_id = ?"
			args = append(args, *filters.CreditCardID)
		}
		if len(filters.CategoryIDs) > 0 {
			placeholders := make([]string, len(filters.CategoryIDs))
			for i, id := range filters.CategoryIDs {
				placeholders[i] = "?"
				args = append(args, id)
			}
			query += fmt.Sprintf(" AND category_id IN (%s)", strings.Join(placeholders, ","))
		}
	}

	query += " ORDER BY realization_date DESC, id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}
	defer rows.Close()

	var entries []*entity.Entry
	for rows.Next() {
		var entry entity.Entry
		var description sql.NullString
		var categoryID, creditCardID sql.NullInt64
		var paymentDate sql.NullString

		if err := rows.Scan(
			&entry.ID,
			&entry.NamespaceID,
			&entry.Type,
			&entry.Amount,
			&entry.Currency,
			&description,
			&categoryID,
			&creditCardID,
			&entry.RealizationDate,
			&paymentDate,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if description.Valid {
			entry.Description = description.String
		}
		if categoryID.Valid {
			entry.CategoryID = &categoryID.Int64
		}
		if creditCardID.Valid {
			entry.CreditCardID = &creditCardID.Int64
		}
		if paymentDate.Valid {
			t, _ := time.Parse("2006-01-02", paymentDate.String)
			entry.PaymentDate = &t
		}

		tags, err := r.getTagsByEntryID(entry.ID)
		if err != nil {
			return nil, err
		}
		entry.Tags = tags

		entries = append(entries, &entry)
	}

	if filters != nil && len(filters.Tags) > 0 {
		entries = r.filterByTags(entries, filters.Tags)
	}

	return entries, nil
}

func (r *EntriesRepository) filterByTags(entries []*entity.Entry, filterTags []string) []*entity.Entry {
	var filtered []*entity.Entry
	for _, entry := range entries {
		for _, ft := range filterTags {
			for _, et := range entry.Tags {
				if et == ft {
					filtered = append(filtered, entry)
					break
				}
			}
			break
		}
	}
	return filtered
}

func (r *EntriesRepository) Update(entry *entity.Entry) error {
	query := `UPDATE entries SET namespace_id = ?, type = ?, amount = ?, currency = ?, description = ?, category_id = ?, credit_card_id = ?, realization_date = ?, payment_date = ? WHERE id = ?`

	var categoryID, creditCardID interface{}
	if entry.CategoryID != nil {
		categoryID = *entry.CategoryID
	}
	if entry.CreditCardID != nil {
		creditCardID = *entry.CreditCardID
	}

	var paymentDate interface{}
	if entry.PaymentDate != nil {
		paymentDate = entry.PaymentDate.Format("2006-01-02")
	}

	_, err := r.db.Exec(query,
		entry.NamespaceID,
		entry.Type,
		entry.Amount,
		entry.Currency,
		entry.Description,
		categoryID,
		creditCardID,
		entry.RealizationDate.Format("2006-01-02"),
		paymentDate,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

func (r *EntriesRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM entry_tags WHERE entry_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry tags: %w", err)
	}

	_, err = r.db.Exec(`DELETE FROM entries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	return nil
}

func (r *EntriesRepository) AddTag(entryID int64, tag string) error {
	query := `INSERT OR IGNORE INTO entry_tags (entry_id, tag) VALUES (?, ?)`
	_, err := r.db.Exec(query, entryID, tag)
	if err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}
	return nil
}

func (r *EntriesRepository) RemoveTag(entryID int64, tag string) error {
	query := `DELETE FROM entry_tags WHERE entry_id = ? AND tag = ?`
	_, err := r.db.Exec(query, entryID, tag)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}
	return nil
}

func (r *EntriesRepository) getTagsByEntryID(entryID int64) ([]string, error) {
	query := `SELECT tag FROM entry_tags WHERE entry_id = ?`

	rows, err := r.db.Query(query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
