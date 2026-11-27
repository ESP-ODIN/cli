package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withRegistry(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	prev := RegistryURL
	RegistryURL = srv.URL
	t.Cleanup(func() { RegistryURL = prev })
}

func TestSearchPackages_MapsResults(t *testing.T) {
	var gotQuery string
	withRegistry(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/packages/search" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		gotQuery = r.URL.Query().Get("q")
		w.Write([]byte(`{"count":1,"data":[{
			"id":"0b6e7d1e-6f43-4c39-9b8f-4d1e0e3d6c11",
			"name":"changelog-agent","category":"dev","agent_type":"cli","runtime":"python",
			"description":"Génère un changelog","is_official_pick":true,"downloads_count":42,
			"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]}`))
	})

	agents, err := SearchPackages("change log")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "change log" {
		t.Errorf("query not forwarded/escaped correctly, got %q", gotQuery)
	}
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	a := agents[0]
	if a.Name != "changelog-agent" || a.Description != "Génère un changelog" || a.Runtime != "python" {
		t.Errorf("bad mapping: %+v", a)
	}
	if !a.IsOfficialPick || a.DownloadsCount != 42 {
		t.Errorf("marketplace metadata not mapped: %+v", a)
	}
}

func TestSearchPackages_EmptyQueryHasNoParam(t *testing.T) {
	withRegistry(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.Write([]byte(`{"count":0,"data":[]}`))
	})

	agents, err := SearchPackages("")
	if err != nil || len(agents) != 0 {
		t.Fatalf("expected empty result, got %v / %v", agents, err)
	}
}

func TestSearchPackages_Errors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   string
	}{
		"api error":  {500, `{"code":"INTERNAL_SERVER_ERROR","message":"boom"}`, "INTERNAL_SERVER_ERROR"},
		"plain body": {502, `Bad Gateway`, "HTTP 502"},
		"bad json":   {200, `not json`, "invalid registry response"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			withRegistry(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			})
			_, err := SearchPackages("x")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}
