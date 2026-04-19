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

func NewEntriesRepository(basePath string) *EntriesRepository {
	return &EntriesRepository{basePath: basePath}
}

func (r *EntriesRepository) filePath(accountName string, year int, month time.Month) string {
	return filepath.Join(r.basePath, accountName,
		fmt.Sprintf("%d", year),
		fmt.Sprintf("%d-%02d-%s-entries.yaml", year, month, accountName))
}

func (r *EntriesRepository) getStorageDate(entry *entity.Entry) time.Time {
	if entry.PaymentDate != nil {
		return *entry.PaymentDate
	}
	return entry.RealizationDate
}

func (r *EntriesRepository) Create(entry *entity.Entry, accountName string) error {
	storageDate := r.getStorageDate(entry)
	path := r.filePath(accountName, storageDate.Year(), storageDate.Month())

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data := &EntriesData{}
	if _, err := os.Stat(path); err == nil {
		readData, err := Read[EntriesData](path)
		if err != nil {
			return err
		}
		data = readData
	}

	yamlEntry := Entry{}
	yamlEntry.FromEntity(entry)
	data.Entries = append(data.Entries, yamlEntry)

	return Write(path, data)
}

func (r *EntriesRepository) GetAll(filters *port.EntryFilters) ([]*entity.Entry, error) {
	accountsData, err := Read[AccountsData](filepath.Join(r.basePath, "accounts.yaml"))
	if err != nil {
		return nil, err
	}

	var entries []*entity.Entry
	currentYear := time.Now().Year()
	currentMonth := time.Now().Month()

	for _, accountName := range accountsData.Accounts {
		if filters != nil && filters.AccountName != "" && filters.AccountName != accountName {
			continue
		}

		for year := 2020; year <= currentYear+1; year++ {
			for month := time.January; month <= time.December; month++ {
				if year == currentYear+1 && month > currentMonth {
					break
				}
				path := r.filePath(accountName, year, month)
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
						if filters.Type != nil && entry.Type != *filters.Type {
							continue
						}
						if filters.FromDate != nil {
							var filterDate time.Time
							if filters.FilterByRealizationDate {
								filterDate = entry.RealizationDate
							} else {
								filterDate = entry.RealizationDate
								if entry.PaymentDate != nil {
									filterDate = *entry.PaymentDate
								}
							}
							if filterDate.Before(*filters.FromDate) {
								continue
							}
						}
						if filters.ToDate != nil {
							var filterDate time.Time
							if filters.FilterByRealizationDate {
								filterDate = entry.RealizationDate
							} else {
								filterDate = entry.RealizationDate
								if entry.PaymentDate != nil {
									filterDate = *entry.PaymentDate
								}
							}
							if filterDate.After(*filters.ToDate) {
								continue
							}
						}
						if filters.CategoryAlias != "" {
							if entry.CategoryAlias == nil || *entry.CategoryAlias != filters.CategoryAlias {
								continue
							}
						}
						if len(filters.Tags) > 0 {
							found := false
							for _, filterTag := range filters.Tags {
								for _, entryTag := range entry.Tags {
									if entryTag == filterTag {
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
						if filters.CreditCard != "" {
							if entry.CreditCardName == nil || *entry.CreditCardName != filters.CreditCard {
								continue
							}
						}
					}

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
			return e1.RealizationDate.Before(e2.RealizationDate)
		}
		if pd1 == nil {
			return e1.RealizationDate.Before(*pd2)
		}
		if pd2 == nil {
			return pd1.Before(e2.RealizationDate)
		}
		if pd1.Equal(*pd2) {
			return e1.RealizationDate.Before(e2.RealizationDate)
		}
		return pd1.Before(*pd2)
	})

	return entries, nil
}

func (r *EntriesRepository) EnsureInitialized() error {
	return nil
}

func (r *EntriesRepository) Port() port.EntriesRepository {
	return r
}
