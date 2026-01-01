package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const AppName = "line-cli"

// ConfigDir returns the config directory path.
// Uses XDG_CONFIG_HOME on Linux, ~/Library/Application Support on macOS.
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", AppName), nil
	}
	return filepath.Join(home, ".config", AppName), nil
}

// DataDir returns the data directory path.
func DataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", AppName), nil
	}
	return filepath.Join(home, ".local", "share", AppName), nil
}

// CacheDir returns the cache directory path.
func CacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Caches", AppName), nil
	}
	return filepath.Join(home, ".cache", AppName), nil
}
