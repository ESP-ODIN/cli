package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// RegistryURL is the base URL of the ODIN backend registry.
// Override at build time: -ldflags "-X github.com/ESP-ODIN/cli/internal.RegistryURL=https://..."
var RegistryURL = "http://localhost:3500"

// Agent represents an installed agent.
type Agent struct {
	Name        string `json:"name"        toml:"name"`
	Version     string `json:"version"     toml:"version"`
	Description string `json:"description" toml:"description"`
	// Repo is the GitHub "owner/repo" where the agent's manifest lives.
	Repo string `json:"repo" toml:"repo"`
	// Run is the shell command to execute the agent.
	Run      string   `json:"run"      toml:"entrypoint"`
	Requires []string `json:"requires,omitempty" toml:"requires"`
}

// Registry is the local list of installed agents persisted at ~/.odin/registry.json.
type Registry struct {
	Agents map[string]Agent `json:"agents"`
}

func getRegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".odin", "registry.json")
}

// AgentDir returns the local directory where an agent's files are stored.
func AgentDir(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".odin", "agents", name)
}

// LoadRegistry reads ~/.odin/registry.json; returns an empty registry if missing.
func LoadRegistry() Registry {
	path := getRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{Agents: make(map[string]Agent)}
	}
	var reg Registry
	json.Unmarshal(data, &reg)
	if reg.Agents == nil {
		reg.Agents = make(map[string]Agent)
	}
	return reg
}

// SaveRegistry writes the registry to disk.
func SaveRegistry(reg Registry) error {
	path := getRegistryPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// FetchAgentList fetches the list of available agents from the backend registry.
func FetchAgentList() ([]Agent, error) {
	resp, err := http.Get(RegistryURL + "/agents")
	if err != nil {
		return nil, fmt.Errorf("cannot reach registry: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	var agents []Agent
	if err := json.Unmarshal(body, &agents); err != nil {
		return nil, fmt.Errorf("invalid registry response: %w", err)
	}
	return agents, nil
}

// FetchAgentByName resolves an agent name via the backend registry, then fetches
// its odin.toml manifest from the agent's GitHub repo.
func FetchAgentByName(name string) (Agent, error) {
	resp, err := http.Get(fmt.Sprintf("%s/agents/%s", RegistryURL, name))
	if err != nil {
		return Agent{}, fmt.Errorf("cannot reach registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return Agent{}, fmt.Errorf("agent '%s' not found in registry", name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Agent{}, fmt.Errorf("read error: %w", err)
	}

	var entry Agent
	if err := json.Unmarshal(body, &entry); err != nil {
		return Agent{}, fmt.Errorf("invalid registry response: %w", err)
	}

	if entry.Repo == "" {
		return Agent{}, fmt.Errorf("agent '%s' has no GitHub repo configured", name)
	}

	return FetchManifestFromGitHub(entry.Repo)
}

// FetchManifestFromGitHub fetches and parses odin.toml (or manifest.toml as fallback)
// from the default branch of the given "owner/repo".
func FetchManifestFromGitHub(ownerRepo string) (Agent, error) {
	for _, filename := range []string{"odin.toml", "manifest.toml"} {
		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", ownerRepo, filename)
		resp, err := http.Get(url)
		if err != nil {
			return Agent{}, fmt.Errorf("cannot fetch manifest from GitHub: %w", err)
		}

		if resp.StatusCode == 404 {
			resp.Body.Close()
			continue
		}

		var agent Agent
		_, decodeErr := toml.NewDecoder(resp.Body).Decode(&agent)
		resp.Body.Close()
		if decodeErr != nil {
			return Agent{}, fmt.Errorf("invalid %s in %s: %w", filename, ownerRepo, decodeErr)
		}

		// manifest.toml uses "entrypoint" but toml tag on Run is already "entrypoint".
		// If Run is still empty, try the "entrypoint" key via a raw decode.
		if agent.Run == "" {
			agent.Run = fetchEntrypoint(ownerRepo, filename)
		}

		agent.Repo = ownerRepo
		return agent, nil
	}

	return Agent{}, fmt.Errorf("no odin.toml or manifest.toml found in %s (branch: main)", ownerRepo)
}

// fetchEntrypoint is a fallback raw decode to read the "entrypoint" key when the
// struct tag mapping doesn't cover it (e.g. manifest.toml uses "entrypoint", not "run").
func fetchEntrypoint(ownerRepo, filename string) string {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", ownerRepo, filename)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	toml.NewDecoder(resp.Body).Decode(&raw)
	if v, ok := raw["entrypoint"].(string); ok {
		return v
	}
	return ""
}
