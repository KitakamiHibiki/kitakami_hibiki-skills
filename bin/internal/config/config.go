package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = ".kitakami_hibiki"

// BaseDir returns the base config directory path: ~/.kitakami_hibiki/config.
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, appDirName, "config"), nil
}

// EnsureDir creates the config directory tree if it does not exist.
func EnsureDir() (string, error) {
	dir, err := BaseDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return dir, nil
}

// Path returns the full path beneath the config directory.
func Path(name string) (string, error) {
	dir, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// AppDir returns the application root directory: ~/.kitakami_hibiki.
func AppDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, appDirName), nil
}

// DownloadDir returns the pixiv download directory, creating it if needed.
func DownloadDir(sub string) (string, error) {
	app, err := AppDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(app, "pixiv", sub)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}
	return dir, nil
}
