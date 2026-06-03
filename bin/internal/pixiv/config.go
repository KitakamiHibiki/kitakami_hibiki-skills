package pixiv

import (
	"encoding/json"
	"fmt"
	"os"

	"skills/bin/internal/config"
)

const configFileName = "pixiv.json"

// Config holds Pixiv credentials persisted to disk.
type Config struct {
	PHPSESSID string `json:"phpsessid"`
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	return config.Path(configFileName)
}

// LoadConfig reads saved credentials from disk. Returns nil if no config exists.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes credentials to disk, creating directories as needed.
func SaveConfig(cfg *Config) error {
	if _, err := config.EnsureDir(); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// ClearConfig deletes the saved config file.
func ClearConfig() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config: %w", err)
	}
	return nil
}
