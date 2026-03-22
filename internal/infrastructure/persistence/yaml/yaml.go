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

func EnsureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
