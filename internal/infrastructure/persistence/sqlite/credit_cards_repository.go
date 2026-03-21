package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CreditCardsRepository struct {
	db *DB
}

func NewCreditCardsRepository(db *DB) *CreditCardsRepository {
	return &CreditCardsRepository{db: db}
}

func (r *CreditCardsRepository) Create(cc *entity.CreditCard) (int64, error) {
	query := `INSERT INTO credit_cards (namespace_id, name, closing_day, due_day) VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, cc.NamespaceID, cc.Name, cc.ClosingDay, cc.DueDay)
	if err != nil {
		return 0, fmt.Errorf("failed to create credit card: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *CreditCardsRepository) GetByID(id int64) (*entity.CreditCard, error) {
	query := `SELECT id, namespace_id, name, closing_day, due_day FROM credit_cards WHERE id = ?`

	var cc entity.CreditCard

	err := r.db.QueryRow(query, id).Scan(&cc.ID, &cc.NamespaceID, &cc.Name, &cc.ClosingDay, &cc.DueDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credit card: %w", err)
	}

	return &cc, nil
}

func (r *CreditCardsRepository) GetByNamespaceID(namespaceID int64) ([]*entity.CreditCard, error) {
	query := `SELECT id, namespace_id, name, closing_day, due_day FROM credit_cards WHERE namespace_id = ? ORDER BY name`

	rows, err := r.db.Query(query, namespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit cards: %w", err)
	}
	defer rows.Close()

	var cards []*entity.CreditCard
	for rows.Next() {
		var cc entity.CreditCard

		if err := rows.Scan(&cc.ID, &cc.NamespaceID, &cc.Name, &cc.ClosingDay, &cc.DueDay); err != nil {
			return nil, fmt.Errorf("failed to scan credit card: %w", err)
		}

		cards = append(cards, &cc)
	}

	return cards, nil
}

func (r *CreditCardsRepository) Update(cc *entity.CreditCard) error {
	query := `UPDATE credit_cards SET namespace_id = ?, name = ?, closing_day = ?, due_day = ? WHERE id = ?`

	_, err := r.db.Exec(query, cc.NamespaceID, cc.Name, cc.ClosingDay, cc.DueDay, cc.ID)
	if err != nil {
		return fmt.Errorf("failed to update credit card: %w", err)
	}

	return nil
}

func (r *CreditCardsRepository) Delete(id int64) error {
	query := `DELETE FROM credit_cards WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete credit card: %w", err)
	}

	return nil
}
