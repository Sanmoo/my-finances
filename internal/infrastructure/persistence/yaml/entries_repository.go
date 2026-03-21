package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type EntriesRepository struct {
	basePath string
}

type EntriesData struct {
	Entries []Entry `yaml:"entries"`
}

func NewEntriesRepository(basePath string) *EntriesRepository {
	return &EntriesRepository{basePath: basePath}
}

func (r *EntriesRepository) filePath(accountName string, year int) string {
	return filepath.Join(r.basePath, fmt.Sprintf("%d-%s-entries.yaml", year, accountName))
}

func (r *EntriesRepository) metaPath() string {
	return filepath.Join(r.basePath, "_meta.yaml")
}

func (r *EntriesRepository) getEntryFilePath(entry *entity.Entry, accounts []Account) (string, error) {
	accountName := ""
	for _, acc := range accounts {
		if acc.ID == entry.AccountID {
			accountName = acc.Name
			break
		}
	}
	if accountName == "" {
		return "", fmt.Errorf("account not found for entry")
	}
	return r.filePath(accountName, entry.RealizationDate.Year()), nil
}

func (r *EntriesRepository) loadEntriesForAccount(accountID int64, year int, accounts []Account) ([]*entity.Entry, error) {
	accountName := ""
	for _, acc := range accounts {
		if acc.ID == accountID {
			accountName = acc.Name
			break
		}
	}
	if accountName == "" {
		return []*entity.Entry{}, nil
	}

	path := r.filePath(accountName, year)
	data, err := Read[EntriesData](path)
	if err != nil {
		return nil, err
	}

	entries := make([]*entity.Entry, 0)
	for i := range data.Entries {
		if data.Entries[i].AccountID == accountID {
			entries = append(entries, data.Entries[i].ToEntity())
		}
	}

	return entries, nil
}

func (r *EntriesRepository) Create(entry *entity.Entry) (int64, error) {
	accounts, err := Read[AccountsData](filepath.Join(r.basePath, "accounts.yaml"))
	if err != nil {
		return 0, err
	}

	accountName := ""
	for _, acc := range accounts.Accounts {
		if acc.ID == entry.AccountID {
			accountName = acc.Name
			break
		}
	}
	if accountName == "" {
		return 0, fmt.Errorf("account not found for entry")
	}

	path := r.filePath(accountName, entry.RealizationDate.Year())

	if err := EnsureMetaFile(r.metaPath()); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(), "entries")
	if err != nil {
		return 0, err
	}

	entry.ID = nextID

	data := &EntriesData{}
	if _, err := os.Stat(path); err == nil {
		readData, err := Read[EntriesData](path)
		if err != nil {
			return 0, err
		}
		data = readData
	}

	yamlEntry := Entry{}
	yamlEntry.FromEntity(entry)
	data.Entries = append(data.Entries, yamlEntry)

	sort.Slice(data.Entries, func(i, j int) bool {
		e1 := data.Entries[i]
		e2 := data.Entries[j]

		var pd1, pd2 *time.Time
		if e1.PaymentDate != nil && *e1.PaymentDate != "" {
			if t, err := time.Parse("2006-01-02", *e1.PaymentDate); err == nil {
				pd1 = &t
			}
		}
		if e2.PaymentDate != nil && *e2.PaymentDate != "" {
			if t, err := time.Parse("2006-01-02", *e2.PaymentDate); err == nil {
				pd2 = &t
			}
		}

		if pd1 == nil && pd2 == nil {
			t1, _ := time.Parse("2006-01-02", e1.RealizationDate)
			t2, _ := time.Parse("2006-01-02", e2.RealizationDate)
			if t1.Equal(t2) {
				return e1.ID < e2.ID
			}
			return t1.Before(t2)
		}
		if pd1 == nil {
			t1, _ := time.Parse("2006-01-02", e1.RealizationDate)
			return t1.Before(*pd2)
		}
		if pd2 == nil {
			t1, _ := time.Parse("2006-01-02", e1.RealizationDate)
			return t1.Before(*pd1)
		}
		if pd1.Equal(*pd2) {
			t1, _ := time.Parse("2006-01-02", e1.RealizationDate)
			t2, _ := time.Parse("2006-01-02", e2.RealizationDate)
			if t1.Equal(t2) {
				return e1.ID < e2.ID
			}
			return t1.Before(t2)
		}
		return pd1.Before(*pd2)
	})

	if err := Write(path, data); err != nil {
		return 0, err
	}

	return entry.ID, nil
}

func (r *EntriesRepository) GetByID(id int64) (*entity.Entry, error) {
	entries, err := r.GetAll(&port.EntryFilters{})
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, nil
}

func (r *EntriesRepository) GetAll(filters *port.EntryFilters) ([]*entity.Entry, error) {
	accounts, err := Read[AccountsData](filepath.Join(r.basePath, "accounts.yaml"))
	if err != nil {
		return nil, err
	}

	var entries []*entity.Entry
	currentYear := time.Now().Year()

	for _, acc := range accounts.Accounts {
		if filters != nil && filters.AccountID != nil && *filters.AccountID != acc.ID {
			continue
		}

		for year := 2020; year <= currentYear+1; year++ {
			path := r.filePath(acc.Name, year)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}

			data, err := Read[EntriesData](path)
			if err != nil {
				continue
			}

			for i := range data.Entries {
				entry := data.Entries[i].ToEntity()
				if filters != nil {
					if filters.AccountID != nil && *filters.AccountID != entry.AccountID {
						continue
					}
					if filters.Type != nil && entry.Type != *filters.Type {
						continue
					}
					if filters.FromDate != nil {
						t, _ := time.Parse("2006-01-02", data.Entries[i].RealizationDate)
						if t.Before(*filters.FromDate) {
							continue
						}
					}
					if filters.ToDate != nil {
						t, _ := time.Parse("2006-01-02", data.Entries[i].RealizationDate)
						if t.After(*filters.ToDate) {
							continue
						}
					}
					if len(filters.CategoryIDs) > 0 {
						found := false
						for _, catID := range filters.CategoryIDs {
							if entry.CategoryID != nil && *entry.CategoryID == catID {
								found = true
								break
							}
						}
						if !found {
							continue
						}
					}
				}

				if filters != nil && len(filters.Tags) > 0 {
					found := false
					for _, ft := range filters.Tags {
						for _, et := range entry.Tags {
							if et == ft {
								found = true
								break
							}
						}
						if found {
							break
						}
					}
					if !found {
						continue
					}
				}

				entries = append(entries, entry)
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		e1 := entries[i]
		e2 := entries[j]

		var pd1, pd2 *time.Time
		if e1.PaymentDate != nil {
			pd1 = e1.PaymentDate
		}
		if e2.PaymentDate != nil {
			pd2 = e2.PaymentDate
		}

		if pd1 == nil && pd2 == nil {
			if e1.RealizationDate.Equal(e2.RealizationDate) {
				return e1.ID < e2.ID
			}
			return e1.RealizationDate.Before(e2.RealizationDate)
		}
		if pd1 == nil {
			return e1.RealizationDate.Before(*pd2)
		}
		if pd2 == nil {
			return e1.RealizationDate.Before(*pd1)
		}
		if pd1.Equal(*pd2) {
			if e1.RealizationDate.Equal(e2.RealizationDate) {
				return e1.ID < e2.ID
			}
			return e1.RealizationDate.Before(e2.RealizationDate)
		}
		return pd1.Before(*pd2)
	})

	return entries, nil
}

func (r *EntriesRepository) GetAllByYear(accountID int64, year int) ([]*entity.Entry, error) {
	accounts, err := Read[AccountsData](filepath.Join(r.basePath, "accounts.yaml"))
	if err != nil {
		return nil, err
	}

	return r.loadEntriesForAccount(accountID, year, accounts.Accounts)
}

func (r *EntriesRepository) Update(entry *entity.Entry) error {
	accounts, err := Read[AccountsData](filepath.Join(r.basePath, "accounts.yaml"))
	if err != nil {
		return err
	}

	path, err := r.getEntryFilePath(entry, accounts.Accounts)
	if err != nil {
		return err
	}

	data, err := Read[EntriesData](path)
	if err != nil {
		return err
	}

	found := false
	for i := range data.Entries {
		if data.Entries[i].ID == entry.ID {
			data.Entries[i].FromEntity(entry)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry not found: %d", entry.ID)
	}

	return Write(path, data)
}

func (r *EntriesRepository) Delete(id int64) error {
	entries, err := r.GetAll(&port.EntryFilters{})
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.ID == id {
			return r.Update(entry)
		}
	}

	return fmt.Errorf("entry not found: %d", id)
}

func (r *EntriesRepository) AddTag(entryID int64, tag string) error {
	entry, err := r.GetByID(entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("entry not found: %d", entryID)
	}

	entry.AddTag(tag)
	return r.Update(entry)
}

func (r *EntriesRepository) RemoveTag(entryID int64, tag string) error {
	entry, err := r.GetByID(entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("entry not found: %d", entryID)
	}

	newTags := make([]string, 0)
	for _, t := range entry.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	entry.Tags = newTags
	return r.Update(entry)
}

func (r *EntriesRepository) EnsureInitialized() error {
	return nil
}

func (r *EntriesRepository) Port() port.EntriesRepository {
	return r
}
