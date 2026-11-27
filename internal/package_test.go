package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serveRegistry(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	prev := RegistryURL
	RegistryURL = srv.URL
	t.Cleanup(func() { RegistryURL = prev })
}

const detailTpl = `{"agent":{"id":"a18a2b10-956f-49c5-8d3e-7e8560f37aea","name":"changelog-agent",
"category":"dev","agent_type":"cli","runtime":"python","description":"desc"},
"version":{"id":"1a97bcbf-b72d-4454-9944-1c767c454451","agent_id":"a18a2b10-956f-49c5-8d3e-7e8560f37aea",
"version":"1.2.0","is_stable":true%s}}`

func TestFetchPackage_UsesEmbeddedManifest(t *testing.T) {
	extra := `,"repo_link":"https://github.com/ESP-ODIN/changelog-agent.git",
"manifest_toml":"name = \"changelog-agent\"\nentrypoint = \"python3 agent.py\"\nrequires = [\"GEMINI_API_KEY\"]\n"`
	serveRegistry(t, strings.Replace(detailTpl, "%s", extra, 1))

	a, err := FetchPackage("changelog-agent", "latest")
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != "1.2.0" || a.Repo != "ESP-ODIN/changelog-agent" || a.Run != "python3 agent.py" {
		t.Errorf("unexpected agent: %+v", a)
	}
	if len(a.Requires) != 1 || a.Requires[0] != "GEMINI_API_KEY" {
		t.Errorf("requires not parsed: %v", a.Requires)
	}
}

func TestFetchPackage_NoManifestNoRepo(t *testing.T) {
	serveRegistry(t, strings.Replace(detailTpl, "%s", "", 1))

	a, err := FetchPackage("changelog-agent", "latest")
	if err != nil {
		t.Fatal(err)
	}
	if a.Run != "" || a.Repo != "" {
		t.Errorf("expected no entrypoint/repo, got %+v", a)
	}
}

func TestFetchPackage_InvalidManifest(t *testing.T) {
	serveRegistry(t, strings.Replace(detailTpl, "%s", `,"manifest_toml":"entrypoint = "`, 1))

	if _, err := FetchPackage("changelog-agent", "latest"); err == nil {
		t.Fatal("expected an error for an invalid manifest")
	}
}

func TestParseGitHubRepo(t *testing.T) {
	cases := map[string]string{
		"ESP-ODIN/agent":                              "ESP-ODIN/agent",
		"https://github.com/ESP-ODIN/agent":           "ESP-ODIN/agent",
		"https://github.com/ESP-ODIN/agent.git":       "ESP-ODIN/agent",
		"https://www.github.com/ESP-ODIN/agent/":      "ESP-ODIN/agent",
		"git@github.com:ESP-ODIN/agent.git":           "ESP-ODIN/agent",
		"https://github.com/ESP-ODIN/agent/tree/main": "ESP-ODIN/agent",
		"https://gitlab.com/ESP-ODIN/agent":           "",
		"not-a-repo":                                  "",
		"":                                            "",
	}
	for in, want := range cases {
		if got := ParseGitHubRepo(in); got != want {
			t.Errorf("ParseGitHubRepo(%q) = %q, want %q", in, got, want)
		}
	}
}
