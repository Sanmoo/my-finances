package port

import (
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type EntriesRepository interface {
	Create(entry *entity.Entry, accountName string) error
	GetAll(filters *EntryFilters) ([]*entity.Entry, error)
}

type EntryFilters struct {
	FromDate      *time.Time
	ToDate        *time.Time
	Type          *entity.EntryType
	CategoryAlias string
	Tags          []string
	CreditCard    string
	AccountName   string
}
