package yaml

import (
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type AccountsData struct {
	Accounts []string `yaml:"accounts"`
}

type CreditCard struct {
	Name       string `yaml:"name"`
	ClosingDay int    `yaml:"closing_day"`
	DueDay     int    `yaml:"due_day"`
}

func (cc *CreditCard) ToEntity() *entity.CreditCard {
	return &entity.CreditCard{
		Name:       cc.Name,
		ClosingDay: cc.ClosingDay,
		DueDay:     cc.DueDay,
	}
}

func (cc *CreditCard) FromEntity(ecc *entity.CreditCard) {
	cc.Name = ecc.Name
	cc.ClosingDay = ecc.ClosingDay
	cc.DueDay = ecc.DueDay
}

type CreditCardsData struct {
	CreditCards []CreditCard `yaml:"credit_cards"`
}

type Tag struct {
	Name string `yaml:"name"`
}

func (t *Tag) ToEntity() *entity.Tag {
	return &entity.Tag{
		Name: t.Name,
	}
}

func (t *Tag) FromEntity(et *entity.Tag) {
	t.Name = et.Name
}

type TagsData struct {
	Tags []Tag `yaml:"tags"`
}

type Category struct {
	Name  string  `yaml:"name"`
	Alias string  `yaml:"alias"`
	Emoji *string `yaml:"emoji,omitempty"`
	Type  string  `yaml:"type"`
}

func (c *Category) ToEntity() *entity.Category {
	return &entity.Category{
		Name:  c.Name,
		Alias: c.Alias,
		Emoji: c.Emoji,
		Type:  entity.CategoryType(c.Type),
	}
}

func (c *Category) FromEntity(ec *entity.Category) {
	c.Name = ec.Name
	c.Alias = ec.Alias
	c.Emoji = ec.Emoji
	c.Type = string(ec.Type)
}

type CategoriesData struct {
	Categories []Category `yaml:"categories"`
}

type Entry struct {
	Type              string   `yaml:"type"`
	Amount            float64  `yaml:"amount"`
	Currency          string   `yaml:"currency"`
	Description       string   `yaml:"description,omitempty"`
	CategoryAlias     *string  `yaml:"category_alias,omitempty"`
	CreditCardName    *string  `yaml:"credit_card_name,omitempty"`
	Tags              []string `yaml:"tags,omitempty"`
	InstallmentNumber int      `yaml:"installment_number,omitempty"`
	InstallmentTotal  int      `yaml:"installment_total,omitempty"`
	RealizationDate   string   `yaml:"realization_date"`
	PaymentDate       *string  `yaml:"payment_date,omitempty"`
	CreatedAt         string   `yaml:"created_at"`
}

func (e *Entry) ToEntity() *entity.Entry {
	entry := &entity.Entry{
		Type:              entity.EntryType(e.Type),
		Amount:            e.Amount,
		Currency:          e.Currency,
		Description:       e.Description,
		CategoryAlias:     e.CategoryAlias,
		CreditCardName:    e.CreditCardName,
		Tags:              e.Tags,
		InstallmentNumber: e.InstallmentNumber,
		InstallmentTotal:  e.InstallmentTotal,
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
	e.Type = string(ee.Type)
	e.Amount = ee.Amount
	e.Currency = ee.Currency
	e.Description = ee.Description
	e.CategoryAlias = ee.CategoryAlias
	e.CreditCardName = ee.CreditCardName
	e.Tags = ee.Tags
	e.InstallmentNumber = ee.InstallmentNumber
	e.InstallmentTotal = ee.InstallmentTotal
	e.RealizationDate = ee.RealizationDate.Format("2006-01-02")

	if ee.PaymentDate != nil {
		pd := ee.PaymentDate.Format("2006-01-02")
		e.PaymentDate = &pd
	}

	e.CreatedAt = ee.CreatedAt.Format("2006-01-02")
}

type EntriesData struct {
	Entries []Entry `yaml:"entries"`
}
