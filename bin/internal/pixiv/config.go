package pixiv

import (
	"fmt"

	"skills/bin/internal/db"
)

const configKeyPHPSESSID = "pixiv_phpsessid"

// Config holds Pixiv credentials persisted to SQLite.
type Config struct {
	PHPSESSID string
}

// LoadConfig reads saved credentials from SQLite. Returns nil if no config exists.
func LoadConfig() (*Config, error) {
	v, err := db.ConfigGet(configKeyPHPSESSID)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if v == "" {
		return nil, nil
	}
	return &Config{PHPSESSID: v}, nil
}

// SaveConfig writes credentials to SQLite.
func SaveConfig(cfg *Config) error {
	return db.ConfigSet(configKeyPHPSESSID, cfg.PHPSESSID)
}

// ClearConfig deletes the saved credentials from SQLite.
func ClearConfig() error {
	return db.ConfigDelete(configKeyPHPSESSID)
}
