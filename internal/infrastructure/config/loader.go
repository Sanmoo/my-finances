package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	DriverSQLite = "sqlite"
	DriverYAML   = "yaml"
)

type Config struct {
	DatabasesPath   string `koanf:"databases.path"`
	DefaultCurrency string `koanf:"default.currency"`
	Locale          string `koanf:"locale"`
	StorageDriver   string `koanf:"storage.driver"`
	DataPath        string `koanf:"data.path"`
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
		return l.defaultConfig(homeDir), nil
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(l.path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg := &Config{
		DatabasesPath:   k.String("databases.path"),
		DefaultCurrency: k.String("default.currency"),
		Locale:          k.String("locale"),
		StorageDriver:   k.String("storage.driver"),
		DataPath:        k.String("data.path"),
	}

	cfg.DataPath = expandPath(cfg.DataPath, homeDir)

	return l.applyDefaults(homeDir, cfg), nil
}

func (l *Loader) defaultConfig(homeDir string) *Config {
	return l.applyDefaults(homeDir, &Config{})
}

func (l *Loader) applyDefaults(homeDir string, cfg *Config) *Config {
	if cfg.DefaultCurrency == "" {
		cfg.DefaultCurrency = "BRL"
	}
	if cfg.Locale == "" {
		cfg.Locale = "pt-BR"
	}
	if cfg.StorageDriver == "" {
		cfg.StorageDriver = DriverSQLite
	}
	if cfg.DataPath == "" {
		cfg.DataPath = filepath.Join(homeDir, ".myfin", "data")
	}
	return cfg
}

func (l *Loader) Save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content := "# myfin configuration\n"
	if cfg.StorageDriver != "" {
		content += fmt.Sprintf("storage.driver: %s\n", cfg.StorageDriver)
	}
	if cfg.DataPath != "" {
		content += fmt.Sprintf("data.path: %s\n", cfg.DataPath)
	}
	if cfg.DefaultCurrency != "" {
		content += fmt.Sprintf("default.currency: %s\n", cfg.DefaultCurrency)
	}
	if cfg.Locale != "" {
		content += fmt.Sprintf("locale: %s\n", cfg.Locale)
	}
	if cfg.DatabasesPath != "" {
		content += fmt.Sprintf("databases.path: %s\n", cfg.DatabasesPath)
	}

	if err := os.WriteFile(l.path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (l *Loader) GetPath() string {
	return l.path
}

func expandPath(path, homeDir string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}
