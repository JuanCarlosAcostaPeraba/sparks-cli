package storage

import (
	"os"
	"path/filepath"
	"runtime"
)

const appDirName = "sparks"

func DefaultDBPath() (string, error) {
	base, err := defaultDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName, "sparks.db"), nil
}

func defaultDataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData, nil
		}
		return os.UserConfigDir()
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return xdg, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}
