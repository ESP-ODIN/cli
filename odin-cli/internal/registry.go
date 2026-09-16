package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Agent représente un agent installé
type Agent struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Run         string `json:"run"`
}

// Registry est la liste de tous les agents installés
type Registry struct {
	Agents map[string]Agent `json:"agents"`
}

// getRegistryPath retourne le chemin vers ~/.odin/registry.json
func getRegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".odin", "registry.json")
}

// LoadRegistry lit le fichier registry.json
// Si il n'existe pas encore, retourne un registry vide
func LoadRegistry() Registry {
	path := getRegistryPath()

	data, err := os.ReadFile(path)
	if err != nil {
		// Fichier pas encore créé → registry vide
		return Registry{Agents: make(map[string]Agent)}
	}

	var reg Registry
	json.Unmarshal(data, &reg)
	return reg
}

// SaveRegistry écrit le registry sur le disque
func SaveRegistry(reg Registry) error {
	path := getRegistryPath()

	// Crée le dossier ~/.odin/ si il n'existe pas
	os.MkdirAll(filepath.Dir(path), 0755)

	// Convertit en JSON indenté (lisible humainement)
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
