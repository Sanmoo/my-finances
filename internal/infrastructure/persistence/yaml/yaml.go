package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Date struct {
	time.Time
}

func (d Date) MarshalYAML() (interface{}, error) {
	return d.Time.Format("2006-01-02"), nil
}

func (d *Date) UnmarshalYAML(value *yaml.Node) error {
	var dateStr string
	if err := value.Decode(&dateStr); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return err
		}
	}
	d.Time = t
	return nil
}

type DateTime struct {
	time.Time
}

func (d DateTime) MarshalYAML() (interface{}, error) {
	return d.Time.Format("2006-01-02"), nil
}

func (d *DateTime) UnmarshalYAML(value *yaml.Node) error {
	var dateStr string
	if err := value.Decode(&dateStr); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return err
		}
	}
	d.Time = t
	return nil
}

type Meta struct {
	NextIDs struct {
		Accounts    int64 `yaml:"accounts"`
		Categories  int64 `yaml:"categories"`
		CreditCards int64 `yaml:"credit_cards"`
		Entries     int64 `yaml:"entries"`
	} `yaml:"next_ids"`
}

var (
	mu   sync.Mutex
	data map[string]*yaml.Node
)

func Init(basePath string) error {
	mu.Lock()
	defer mu.Unlock()

	data = make(map[string]*yaml.Node)

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := loadFile(filepath.Join(basePath, "_meta.yaml")); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func loadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return fmt.Errorf("failed to parse yaml %s: %w", path, err)
	}

	data[path] = &node
	return nil
}

func Read[T any](path string) (*T, error) {
	mu.Lock()
	defer mu.Unlock()

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			var empty T
			return &empty, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result T
	if err := yaml.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return &result, nil
}

func Write[T any](path string, data T) error {
	mu.Lock()
	defer mu.Unlock()

	content, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func Append[T any](path string, item T) error {
	mu.Lock()
	defer mu.Unlock()

	var items []T
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(content) > 0 {
		if err := yaml.Unmarshal(content, &items); err != nil {
			return fmt.Errorf("failed to unmarshal existing yaml: %w", err)
		}
	}

	items = append(items, item)

	newContent, err := yaml.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func GetNextID(metaPath string, entity string) (int64, error) {
	mu.Lock()
	defer mu.Unlock()

	var meta Meta
	content, err := os.ReadFile(metaPath)
	if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("failed to read meta file: %w", err)
	}

	if len(content) > 0 {
		if err := yaml.Unmarshal(content, &meta); err != nil {
			return 0, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	var nextID int64
	switch entity {
	case "accounts":
		nextID = meta.NextIDs.Accounts
		meta.NextIDs.Accounts++
	case "categories":
		nextID = meta.NextIDs.Categories
		meta.NextIDs.Categories++
	case "credit_cards":
		nextID = meta.NextIDs.CreditCards
		meta.NextIDs.CreditCards++
	case "entries":
		nextID = meta.NextIDs.Entries
		meta.NextIDs.Entries++
	default:
		return 0, fmt.Errorf("unknown entity: %s", entity)
	}

	newContent, err := yaml.Marshal(meta)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal meta: %w", err)
	}

	if err := os.WriteFile(metaPath, newContent, 0644); err != nil {
		return 0, fmt.Errorf("failed to write meta file: %w", err)
	}

	return nextID, nil
}

func EnsureMetaFile(metaPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		meta := Meta{}
		meta.NextIDs.Accounts = 1
		meta.NextIDs.Categories = 1
		meta.NextIDs.CreditCards = 1
		meta.NextIDs.Entries = 1

		content, err := yaml.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}

		if err := os.WriteFile(metaPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write meta file: %w", err)
		}
	}

	return nil
}
