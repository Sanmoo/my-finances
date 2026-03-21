package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AccountsRepository struct {
	db *DB
}

func NewAccountsRepository(db *DB) *AccountsRepository {
	return &AccountsRepository{db: db}
}

func (r *AccountsRepository) Create(acc *entity.Account) (int64, error) {
	query := `INSERT INTO accounts (namespace_id, name) VALUES (?, ?)`

	result, err := r.db.Exec(query, acc.NamespaceID, acc.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to create account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *AccountsRepository) GetByID(id int64) (*entity.Account, error) {
	query := `SELECT id, namespace_id, name FROM accounts WHERE id = ?`

	var acc entity.Account

	err := r.db.QueryRow(query, id).Scan(&acc.ID, &acc.NamespaceID, &acc.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &acc, nil
}

func (r *AccountsRepository) GetByNamespaceID(namespaceID int64) ([]*entity.Account, error) {
	query := `SELECT id, namespace_id, name FROM accounts WHERE namespace_id = ? ORDER BY name`

	rows, err := r.db.Query(query, namespaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*entity.Account
	for rows.Next() {
		var acc entity.Account

		if err := rows.Scan(&acc.ID, &acc.NamespaceID, &acc.Name); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		accounts = append(accounts, &acc)
	}

	return accounts, nil
}

func (r *AccountsRepository) Update(acc *entity.Account) error {
	query := `UPDATE accounts SET namespace_id = ?, name = ? WHERE id = ?`

	_, err := r.db.Exec(query, acc.NamespaceID, acc.Name, acc.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

func (r *AccountsRepository) Delete(id int64) error {
	query := `DELETE FROM accounts WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}
