package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type NamespacesRepository struct {
	db *DB
}

func NewNamespacesRepository(db *DB) *NamespacesRepository {
	return &NamespacesRepository{db: db}
}

func (r *NamespacesRepository) Create(ns *entity.Namespace) (int64, error) {
	query := `INSERT INTO namespaces (name, default_credit_card_id, default_currency) VALUES (?, ?, ?)`

	var defaultCCID interface{}
	if ns.DefaultCreditCardID != nil {
		defaultCCID = *ns.DefaultCreditCardID
	}

	result, err := r.db.Exec(query, ns.Name, defaultCCID, ns.DefaultCurrency)
	if err != nil {
		return 0, fmt.Errorf("failed to create namespace: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *NamespacesRepository) GetByID(id int64) (*entity.Namespace, error) {
	query := `SELECT id, name, default_credit_card_id, default_currency FROM namespaces WHERE id = ?`

	var ns entity.Namespace
	var defaultCCID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(&ns.ID, &ns.Name, &defaultCCID, &ns.DefaultCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if defaultCCID.Valid {
		ns.DefaultCreditCardID = &defaultCCID.Int64
	}

	return &ns, nil
}

func (r *NamespacesRepository) GetByName(name string) (*entity.Namespace, error) {
	query := `SELECT id, name, default_credit_card_id, default_currency FROM namespaces WHERE name = ?`

	var ns entity.Namespace
	var defaultCCID sql.NullInt64

	err := r.db.QueryRow(query, name).Scan(&ns.ID, &ns.Name, &defaultCCID, &ns.DefaultCurrency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if defaultCCID.Valid {
		ns.DefaultCreditCardID = &defaultCCID.Int64
	}

	return &ns, nil
}

func (r *NamespacesRepository) GetAll() ([]*entity.Namespace, error) {
	query := `SELECT id, name, default_credit_card_id, default_currency FROM namespaces ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []*entity.Namespace
	for rows.Next() {
		var ns entity.Namespace
		var defaultCCID sql.NullInt64

		if err := rows.Scan(&ns.ID, &ns.Name, &defaultCCID, &ns.DefaultCurrency); err != nil {
			return nil, fmt.Errorf("failed to scan namespace: %w", err)
		}

		if defaultCCID.Valid {
			ns.DefaultCreditCardID = &defaultCCID.Int64
		}

		namespaces = append(namespaces, &ns)
	}

	return namespaces, nil
}

func (r *NamespacesRepository) Update(ns *entity.Namespace) error {
	query := `UPDATE namespaces SET name = ?, default_credit_card_id = ?, default_currency = ? WHERE id = ?`

	var defaultCCID interface{}
	if ns.DefaultCreditCardID != nil {
		defaultCCID = *ns.DefaultCreditCardID
	}

	_, err := r.db.Exec(query, ns.Name, defaultCCID, ns.DefaultCurrency, ns.ID)
	if err != nil {
		return fmt.Errorf("failed to update namespace: %w", err)
	}

	return nil
}

func (r *NamespacesRepository) Delete(id int64) error {
	query := `DELETE FROM namespaces WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}
