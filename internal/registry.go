package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

// RegistryURL is the base URL of the ODIN marketplace backend.
// Override at build time: -ldflags "-X github.com/ESP-ODIN/cli/internal.RegistryURL=https://..."
var RegistryURL = "http://localhost:8081"

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
	// Fields from the marketplace API
	ID        string `json:"id,omitempty"`
	Category  string `json:"category,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	Runtime   string `json:"runtime,omitempty"`
}

// Registry is the local list of installed agents persisted at ~/.odin/registry.json.
type Registry struct {
	Agents map[string]Agent `json:"agents"`
}

// --- backend DTO types (mirror marketplace models) ---

type backendAgent struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Category       string    `json:"category"`
	AgentType      string    `json:"agent_type"`
	Runtime        string    `json:"runtime"`
	Description    *string   `json:"description"`
	IsOfficialPick bool      `json:"is_official_pick"`
	DownloadsCount int       `json:"downloads_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type backendAgentVersion struct {
	ID        uuid.UUID `json:"id"`
	AgentID   uuid.UUID `json:"agent_id"`
	Version   string    `json:"version"`
	IsStable  bool      `json:"is_stable"`
	RepoLink  *string   `json:"repo_link"`
	Changelog *string   `json:"changelog"`
	// ManifestToml is the raw odin.toml published with this version.
	ManifestToml *string `json:"manifest_toml"`
}

type packageSearchResponse struct {
	Count int            `json:"count"`
	Data  []backendAgent `json:"data"`
}

type packageDetailResponse struct {
	Agent   backendAgent        `json:"agent"`
	Version backendAgentVersion `json:"version"`
}

type cliErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// --- local registry helpers ---

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

// --- marketplace API calls ---

func doGet(path string) ([]byte, int, error) {
	resp, err := http.Get(RegistryURL + path)
	if err != nil {
		return nil, 0, fmt.Errorf("impossible de contacter le registry : %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read error: %w", err)
	}
	return body, resp.StatusCode, nil
}

// SearchPackages calls GET /api/v1/packages/search?q=<query> and returns matching agents.
func SearchPackages(query string) ([]Agent, error) {
	path := "/api/v1/packages/search"
	if query != "" {
		path += "?q=" + url.QueryEscape(query)
	}

	body, status, err := doGet(path)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		var apiErr cliErrorResponse
		json.Unmarshal(body, &apiErr)
		return nil, fmt.Errorf("registry error (%s): %s", apiErr.Code, apiErr.Message)
	}

	var resp packageSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("invalid registry response: %w", err)
	}

	agents := make([]Agent, len(resp.Data))
	for i, a := range resp.Data {
		agents[i] = backendAgentToLocal(a, "")
	}
	return agents, nil
}

// FetchPackage calls GET /api/v1/packages/{name}/{version} and returns the agent ready to install.
// Pass "latest" as version to get the latest stable release.
func FetchPackage(name, version string) (Agent, error) {
	path := fmt.Sprintf("/api/v1/packages/%s/%s", url.PathEscape(name), url.PathEscape(version))

	body, status, err := doGet(path)
	if err != nil {
		return Agent{}, err
	}

	if status == http.StatusNotFound {
		var apiErr cliErrorResponse
		json.Unmarshal(body, &apiErr)
		switch apiErr.Code {
		case "PACKAGE_NOT_FOUND":
			return Agent{}, fmt.Errorf("agent '%s' introuvable dans le registry", name)
		case "VERSION_NOT_FOUND":
			return Agent{}, fmt.Errorf("version '%s' introuvable pour l'agent '%s'", version, name)
		}
		return Agent{}, fmt.Errorf("agent '%s@%s' introuvable", name, version)
	}

	if status != http.StatusOK {
		var apiErr cliErrorResponse
		json.Unmarshal(body, &apiErr)
		return Agent{}, fmt.Errorf("registry error (%s): %s", apiErr.Code, apiErr.Message)
	}

	var detail packageDetailResponse
	if err := json.Unmarshal(body, &detail); err != nil {
		return Agent{}, fmt.Errorf("invalid registry response: %w", err)
	}

	agent := backendAgentToLocal(detail.Agent, detail.Version.Version)
	if detail.Version.RepoLink != nil {
		agent.Repo = ParseGitHubRepo(*detail.Version.RepoLink)
	}

	// Resolve how to run the agent: prefer the manifest published with the
	// version, fall back to the odin.toml hosted in the agent's GitHub repo.
	var manifest Agent
	var manifestErr error
	if detail.Version.ManifestToml != nil && strings.TrimSpace(*detail.Version.ManifestToml) != "" {
		manifest, manifestErr = ParseManifest(*detail.Version.ManifestToml)
	} else if agent.Repo != "" {
		manifest, manifestErr = FetchManifestFromGitHub(agent.Repo)
	}
	if manifestErr != nil {
		return Agent{}, manifestErr
	}
	agent.Run = manifest.Run
	agent.Requires = manifest.Requires

	return agent, nil
}

// ParseManifest decodes an odin.toml manifest.
func ParseManifest(raw string) (Agent, error) {
	var agent Agent
	if _, err := toml.Decode(raw, &agent); err != nil {
		return Agent{}, fmt.Errorf("manifest odin.toml invalide : %w", err)
	}
	return agent, nil
}

// ParseGitHubRepo extracts "owner/repo" from a GitHub link. It accepts
// "owner/repo", "https://github.com/owner/repo(.git)" and
// "git@github.com:owner/repo.git". Returns "" if the link is not recognised.
func ParseGitHubRepo(link string) string {
	s := strings.TrimSpace(link)
	for _, prefix := range []string{"https://", "http://", "git@"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.TrimPrefix(s, "www.")
	if rest, ok := strings.CutPrefix(s, "github.com"); ok {
		s = strings.TrimLeft(rest, ":/")
	} else if strings.Contains(s, ".") && strings.Contains(strings.SplitN(s, "/", 2)[0], ".") {
		return "" // another host
	}
	s = strings.TrimSuffix(strings.TrimSuffix(s, "/"), ".git")

	parts := strings.Split(s, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "/" + strings.TrimSuffix(parts[1], ".git")
}

func backendAgentToLocal(a backendAgent, version string) Agent {
	desc := ""
	if a.Description != nil {
		desc = *a.Description
	}
	return Agent{
		ID:          a.ID.String(),
		Name:        a.Name,
		Version:     version,
		Description: desc,
		Category:    a.Category,
		AgentType:   a.AgentType,
		Runtime:     a.Runtime,
	}
}

// FetchManifestFromGitHub fetches and parses odin.toml (or manifest.toml as fallback)
// from the default branch of the given "owner/repo".
func FetchManifestFromGitHub(ownerRepo string) (Agent, error) {
	for _, filename := range []string{"odin.toml", "manifest.toml"} {
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", ownerRepo, filename)
		resp, err := http.Get(rawURL)
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

		if agent.Run == "" {
			agent.Run = fetchEntrypoint(ownerRepo, filename)
		}

		agent.Repo = ownerRepo
		return agent, nil
	}

	return Agent{}, fmt.Errorf("no odin.toml or manifest.toml found in %s (branch: main)", ownerRepo)
}

func fetchEntrypoint(ownerRepo, filename string) string {
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", ownerRepo, filename)
	resp, err := http.Get(rawURL)
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
