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

func (r *EntriesRepository) filePath(accountName string, year int, month time.Month) string {
	return filepath.Join(r.basePath, accountName, fmt.Sprintf("%d", year),
		fmt.Sprintf("%d-%02d-%s-entries.yaml", year, month, accountName))
}

func (r *EntriesRepository) metaPath(accountName string) string {
	return filepath.Join(r.basePath, accountName, "_meta.yaml")
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
	return r.filePath(accountName, entry.RealizationDate.Year(), entry.RealizationDate.Month()), nil
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

	path := r.filePath(accountName, entry.RealizationDate.Year(), entry.RealizationDate.Month())

	if err := EnsureMetaFile(r.metaPath(accountName)); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(accountName), "entries")
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

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

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
	currentMonth := time.Now().Month()

	for _, acc := range accounts.Accounts {
		if filters != nil && filters.AccountID != nil && *filters.AccountID != acc.ID {
			continue
		}

		for year := 2020; year <= currentYear+1; year++ {
			for month := time.January; month <= time.December; month++ {
				if year == currentYear+1 && month > currentMonth {
					break
				}
				path := r.filePath(acc.Name, year, month)
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

					// Tag filtering disabled - now uses TagIDs instead of tag names
					// TODO: Implement tag ID filtering if needed

					entries = append(entries, entry)
				}
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

func (r *EntriesRepository) AddTag(entryID int64, tagID int64) error {
	entry, err := r.GetByID(entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("entry not found: %d", entryID)
	}

	entry.AddTagID(tagID)
	return r.Update(entry)
}

func (r *EntriesRepository) RemoveTag(entryID int64, tagID int64) error {
	entry, err := r.GetByID(entryID)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("entry not found: %d", entryID)
	}

	newTagIDs := make([]int64, 0)
	for _, id := range entry.TagIDs {
		if id != tagID {
			newTagIDs = append(newTagIDs, id)
		}
	}
	entry.TagIDs = newTagIDs
	return r.Update(entry)
}

func (r *EntriesRepository) EnsureInitialized() error {
	return nil
}

func (r *EntriesRepository) Port() port.EntriesRepository {
	return r
}
