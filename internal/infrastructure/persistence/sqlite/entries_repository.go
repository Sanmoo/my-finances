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
	query := `INSERT INTO entries (type, amount, currency, description, category_id, credit_card_id, account_id, installment, parent_entry_id, realization_date, payment_date, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var categoryID, creditCardID, parentEntryID interface{}
	if entry.CategoryID != nil {
		categoryID = *entry.CategoryID
	}
	if entry.CreditCardID != nil {
		creditCardID = *entry.CreditCardID
	}
	if entry.ParentEntryID != nil {
		parentEntryID = *entry.ParentEntryID
	}

	var paymentDate interface{}
	if entry.PaymentDate != nil {
		paymentDate = entry.PaymentDate.Format("2006-01-02")
	}

	result, err := r.db.Exec(query,
		entry.Type,
		entry.Amount,
		entry.Currency,
		entry.Description,
		categoryID,
		creditCardID,
		entry.AccountID,
		entry.Installment,
		parentEntryID,
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

	for _, tagID := range entry.TagIDs {
		if err := r.addTagID(id, tagID); err != nil {
			return 0, fmt.Errorf("failed to add tag ID: %w", err)
		}
	}

	return id, nil
}

func (r *EntriesRepository) GetByID(id int64) (*entity.Entry, error) {
	query := `SELECT id, type, amount, currency, description, category_id, credit_card_id, account_id, installment, parent_entry_id, realization_date, payment_date, created_at 
			  FROM entries WHERE id = ?`

	var entry entity.Entry
	var description sql.NullString
	var categoryID, creditCardID, parentEntryID sql.NullInt64
	var paymentDate sql.NullString
	var realizationDate, createdAt string

	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.Type,
		&entry.Amount,
		&entry.Currency,
		&description,
		&categoryID,
		&creditCardID,
		&entry.AccountID,
		&entry.Installment,
		&parentEntryID,
		&realizationDate,
		&paymentDate,
		&createdAt,
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
	if parentEntryID.Valid {
		entry.ParentEntryID = &parentEntryID.Int64
	}
	if realizationDate != "" {
		entry.RealizationDate, _ = time.Parse("2006-01-02", realizationDate)
	}
	if paymentDate.Valid {
		t, _ := time.Parse("2006-01-02", paymentDate.String)
		entry.PaymentDate = &t
	}
	if createdAt != "" {
		entry.CreatedAt, _ = time.Parse("2006-01-02", createdAt)
	}

	tagIDs, err := r.getTagIDsByEntryID(entry.ID)
	if err != nil {
		return nil, err
	}
	entry.TagIDs = tagIDs

	return &entry, nil
}

func (r *EntriesRepository) GetAll(filters *port.EntryFilters) ([]*entity.Entry, error) {
	query := `SELECT id, type, amount, currency, description, category_id, credit_card_id, account_id, installment, parent_entry_id, realization_date, payment_date, created_at 
			  FROM entries WHERE 1=1`
	var args []interface{}

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
		if filters.AccountID != nil {
			query += " AND account_id = ?"
			args = append(args, *filters.AccountID)
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

	query += " ORDER BY payment_date ASC, realization_date ASC, id ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}
	defer rows.Close()

	var entries []*entity.Entry
	for rows.Next() {
		var entry entity.Entry
		var description sql.NullString
		var categoryID, creditCardID, parentEntryID sql.NullInt64
		var paymentDate sql.NullString
		var realizationDate, createdAt string

		if err := rows.Scan(
			&entry.ID,
			&entry.Type,
			&entry.Amount,
			&entry.Currency,
			&description,
			&categoryID,
			&creditCardID,
			&entry.AccountID,
			&entry.Installment,
			&parentEntryID,
			&realizationDate,
			&paymentDate,
			&createdAt,
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
		if parentEntryID.Valid {
			entry.ParentEntryID = &parentEntryID.Int64
		}
		if realizationDate != "" {
			entry.RealizationDate, _ = time.Parse("2006-01-02", realizationDate)
		}
		if paymentDate.Valid {
			t, _ := time.Parse("2006-01-02", paymentDate.String)
			entry.PaymentDate = &t
		}
		if createdAt != "" {
			entry.CreatedAt, _ = time.Parse("2006-01-02", createdAt)
		}

		tagIDs, err := r.getTagIDsByEntryID(entry.ID)
		if err != nil {
			return nil, err
		}
		entry.TagIDs = tagIDs

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *EntriesRepository) filterByTagIDs(entries []*entity.Entry, filterTagIDs []int64) []*entity.Entry {
	var filtered []*entity.Entry
	for _, entry := range entries {
		for _, ftID := range filterTagIDs {
			for _, etID := range entry.TagIDs {
				if etID == ftID {
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
	query := `UPDATE entries SET type = ?, amount = ?, currency = ?, description = ?, category_id = ?, credit_card_id = ?, account_id = ?, installment = ?, parent_entry_id = ?, realization_date = ?, payment_date = ? WHERE id = ?`

	var categoryID, creditCardID, parentEntryID interface{}
	if entry.CategoryID != nil {
		categoryID = *entry.CategoryID
	}
	if entry.CreditCardID != nil {
		creditCardID = *entry.CreditCardID
	}
	if entry.ParentEntryID != nil {
		parentEntryID = *entry.ParentEntryID
	}

	var paymentDate interface{}
	if entry.PaymentDate != nil {
		paymentDate = entry.PaymentDate.Format("2006-01-02")
	}

	_, err := r.db.Exec(query,
		entry.Type,
		entry.Amount,
		entry.Currency,
		entry.Description,
		categoryID,
		creditCardID,
		entry.AccountID,
		entry.Installment,
		parentEntryID,
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
	_, err := r.db.Exec(`DELETE FROM entry_tag_ids WHERE entry_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry tag IDs: %w", err)
	}

	_, err = r.db.Exec(`DELETE FROM entries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	return nil
}

func (r *EntriesRepository) addTagID(entryID int64, tagID int64) error {
	query := `INSERT OR IGNORE INTO entry_tag_ids (entry_id, tag_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, entryID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag ID: %w", err)
	}
	return nil
}

func (r *EntriesRepository) getTagIDsByEntryID(entryID int64) ([]int64, error) {
	query := `SELECT tag_id FROM entry_tag_ids WHERE entry_id = ?`

	rows, err := r.db.Query(query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag IDs: %w", err)
	}
	defer rows.Close()

	var tagIDs []int64
	for rows.Next() {
		var tagID int64
		if err := rows.Scan(&tagID); err != nil {
			return nil, fmt.Errorf("failed to scan tag ID: %w", err)
		}
		tagIDs = append(tagIDs, tagID)
	}

	return tagIDs, nil
}

func (r *EntriesRepository) AddTag(entryID int64, tagID int64) error {
	return r.addTagID(entryID, tagID)
}

func (r *EntriesRepository) RemoveTag(entryID int64, tagID int64) error {
	query := `DELETE FROM entry_tag_ids WHERE entry_id = ? AND tag_id = ?`
	_, err := r.db.Exec(query, entryID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag ID: %w", err)
	}
	return nil
}
