// Package config manages winding's list of tracked repos, stored as a
// human-editable YAML file under the user's config directory.
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Dir returns the directory winding's config file lives in: $XDG_CONFIG_HOME/winding
// if set, otherwise ~/.config/winding. os.UserConfigDir() is deliberately not used
// here — on macOS it resolves to ~/Library/Application Support, not ~/.config.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "winding"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "winding"), nil
}

// Path returns the full path to winding's config.yaml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func load() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	path, err := Path()
	if err != nil {
		return nil, err
	}
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return v, nil
}

func save(v *viper.Viper) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return v.WriteConfig()
}

// Repos returns the tracked repo paths, in the order they were added.
func Repos() ([]string, error) {
	v, err := load()
	if err != nil {
		return nil, err
	}
	return v.GetStringSlice("repos"), nil
}

// AddRepo tracks path, reporting whether it was newly added (false if
// already tracked).
func AddRepo(path string) (bool, error) {
	v, err := load()
	if err != nil {
		return false, err
	}

	repos := v.GetStringSlice("repos")
	for _, p := range repos {
		if p == path {
			return false, nil
		}
	}

	v.Set("repos", append(repos, path))
	if err := save(v); err != nil {
		return false, err
	}
	return true, nil
}

// RemoveRepo untracks path, reporting whether it was present.
func RemoveRepo(path string) (bool, error) {
	v, err := load()
	if err != nil {
		return false, err
	}

	repos := v.GetStringSlice("repos")
	remaining := repos[:0]
	removed := false
	for _, p := range repos {
		if p == path {
			removed = true
			continue
		}
		remaining = append(remaining, p)
	}
	if !removed {
		return false, nil
	}

	v.Set("repos", remaining)
	if err := save(v); err != nil {
		return false, err
	}
	return true, nil
}
