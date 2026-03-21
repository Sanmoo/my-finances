package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	DefaultNamespace  string `koanf:"default.namespace"`
	DefaultCurrency   string `koanf:"default.currency"`
	DefaultCreditCard string `koanf:"default.credit-card"`
}

type Loader struct {
	path string
}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	l.path = filepath.Join(homeDir, ".myfin.yaml")

	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		return &Config{
			DefaultCurrency: "BRL",
		}, nil
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(l.path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg := &Config{
		DefaultNamespace:  k.String("default.namespace"),
		DefaultCurrency:   k.String("default.currency"),
		DefaultCreditCard: k.String("default.credit-card"),
	}

	if cfg.DefaultCurrency == "" {
		cfg.DefaultCurrency = "BRL"
	}

	return cfg, nil
}

func (l *Loader) Save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content := "# myfin configuration\n"
	if cfg.DefaultNamespace != "" {
		content += fmt.Sprintf("default.namespace: %s\n", cfg.DefaultNamespace)
	}
	if cfg.DefaultCurrency != "" {
		content += fmt.Sprintf("default.currency: %s\n", cfg.DefaultCurrency)
	}
	if cfg.DefaultCreditCard != "" {
		content += fmt.Sprintf("default.credit-card: %s\n", cfg.DefaultCreditCard)
	}

	if err := os.WriteFile(l.path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
