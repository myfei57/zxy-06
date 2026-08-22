package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// EnsureDir creates dir and all parents when missing.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// WriteFileAtomic writes data to a temporary file and renames it over path.
func WriteFileAtomic(path string, data []byte) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SaveJSON persists a value with an atomic rename.
func SaveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, data)
}

// LoadJSON reads a JSON file; a missing file is not an error.
func LoadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, value)
}
