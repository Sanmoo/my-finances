package yaml

import (
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"time"
)

type Account entity.Account

type Category struct {
	ID    int64   `yaml:"id"`
	Name  string  `yaml:"name"`
	Alias *string `yaml:"alias,omitempty"`
	Emoji *string `yaml:"emoji,omitempty"`
	Type  string  `yaml:"type"`
}

func (c *Category) ToEntity() *entity.Category {
	return &entity.Category{
		ID:    c.ID,
		Name:  c.Name,
		Alias: c.Alias,
		Emoji: c.Emoji,
		Type:  entity.CategoryType(c.Type),
	}
}

func (c *Category) FromEntity(ec *entity.Category) {
	c.ID = ec.ID
	c.Name = ec.Name
	c.Alias = ec.Alias
	c.Emoji = ec.Emoji
	c.Type = string(ec.Type)
}

type CreditCard struct {
	ID         int64  `yaml:"id"`
	Name       string `yaml:"name"`
	ClosingDay int    `yaml:"closing_day"`
	DueDay     int    `yaml:"due_day"`
}

func (cc *CreditCard) ToEntity() *entity.CreditCard {
	return &entity.CreditCard{
		ID:         cc.ID,
		Name:       cc.Name,
		ClosingDay: cc.ClosingDay,
		DueDay:     cc.DueDay,
	}
}

func (cc *CreditCard) FromEntity(ecc *entity.CreditCard) {
	cc.ID = ecc.ID
	cc.Name = ecc.Name
	cc.ClosingDay = ecc.ClosingDay
	cc.DueDay = ecc.DueDay
}

type Entry struct {
	ID              int64    `yaml:"id"`
	Type            string   `yaml:"type"`
	Amount          float64  `yaml:"amount"`
	Currency        string   `yaml:"currency"`
	Description     string   `yaml:"description,omitempty"`
	CategoryID      *int64   `yaml:"category_id,omitempty"`
	CreditCardID    *int64   `yaml:"credit_card_id,omitempty"`
	AccountID       int64    `yaml:"account_id"`
	Installment     int      `yaml:"installment"`
	ParentEntryID   *int64   `yaml:"parent_entry_id,omitempty"`
	RealizationDate string   `yaml:"realization_date"`
	PaymentDate     *string  `yaml:"payment_date,omitempty"`
	CreatedAt       string   `yaml:"created_at"`
	Tags            []string `yaml:"tags,omitempty"`
}

func (e *Entry) ToEntity() *entity.Entry {
	entry := &entity.Entry{
		ID:            e.ID,
		Type:          entity.EntryType(e.Type),
		Amount:        e.Amount,
		Currency:      e.Currency,
		Description:   e.Description,
		CategoryID:    e.CategoryID,
		CreditCardID:  e.CreditCardID,
		AccountID:     e.AccountID,
		Installment:   e.Installment,
		ParentEntryID: e.ParentEntryID,
		Tags:          e.Tags,
	}

	if e.RealizationDate != "" {
		if t, err := time.Parse("2006-01-02", e.RealizationDate); err == nil {
			entry.RealizationDate = t
		}
	}

	if e.PaymentDate != nil && *e.PaymentDate != "" {
		if t, err := time.Parse("2006-01-02", *e.PaymentDate); err == nil {
			entry.PaymentDate = &t
		}
	}

	if e.CreatedAt != "" {
		if t, err := time.Parse("2006-01-02", e.CreatedAt); err == nil {
			entry.CreatedAt = t
		}
	}

	return entry
}

func (e *Entry) FromEntity(ee *entity.Entry) {
	e.ID = ee.ID
	e.Type = string(ee.Type)
	e.Amount = ee.Amount
	e.Currency = ee.Currency
	e.Description = ee.Description
	e.CategoryID = ee.CategoryID
	e.CreditCardID = ee.CreditCardID
	e.AccountID = ee.AccountID
	e.Installment = ee.Installment
	e.ParentEntryID = ee.ParentEntryID
	e.Tags = ee.Tags
	e.RealizationDate = ee.RealizationDate.Format("2006-01-02")

	if ee.PaymentDate != nil {
		pd := ee.PaymentDate.Format("2006-01-02")
		e.PaymentDate = &pd
	}

	e.CreatedAt = ee.CreatedAt.Format("2006-01-02")
}
