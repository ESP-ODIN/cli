package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getKeysPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".odin", "keys.json")
}

func LoadKeys() map[string]string {
	data, err := os.ReadFile(getKeysPath())
	if err != nil {
		return make(map[string]string)
	}
	var keys map[string]string
	json.Unmarshal(data, &keys)
	return keys
}

func SaveKeys(keys map[string]string) error {
	path := getKeysPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
