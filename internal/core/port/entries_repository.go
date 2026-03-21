package port

import (
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type EntriesRepository interface {
	Create(entry *entity.Entry) (int64, error)
	GetByID(id int64) (*entity.Entry, error)
	GetAll(filters *EntryFilters) ([]*entity.Entry, error)
	GetAllByYear(accountID int64, year int) ([]*entity.Entry, error)
	Update(entry *entity.Entry) error
	Delete(id int64) error
	AddTag(entryID int64, tag string) error
	RemoveTag(entryID int64, tag string) error
}

type EntryFilters struct {
	FromDate     *time.Time
	ToDate       *time.Time
	Type         *entity.EntryType
	CategoryIDs  []int64
	Tags         []string
	CreditCardID *int64
	AccountID    *int64
}
