package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMakeAbsolute(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"agent.py", "run.sh"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	abs := func(f string) string { return shellQuote(filepath.Join(dir, f)) }

	cases := map[string]string{
		"python3 agent.py":        "python3 " + abs("agent.py"),
		"python3 agent.py --flag": "python3 " + abs("agent.py") + " --flag",
		"run.sh --x":              abs("run.sh") + " --x",
		"python3 -m mypkg":        "python3 -m mypkg",
		"python3 /abs/agent.py":   "python3 /abs/agent.py",
		"docker run image:latest": "docker run image:latest",
	}
	// filepath.Join cleans "./run.sh" so it resolves too.
	cases["./run.sh"] = abs("run.sh")

	for in, want := range cases {
		if got := makeAbsolute(in, dir); got != want {
			t.Errorf("makeAbsolute(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildCommand_ForwardsArgsVerbatim(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-specific")
	}
	args := []string{"--dry-run", "with space", "$HOME", "a;echo pwned"}
	out, err := buildCommand(`printf '%s\n'`, args).Output()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if strings.Join(got, "|") != strings.Join(args, "|") {
		t.Errorf("args not forwarded verbatim: %q", got)
	}
}

func TestMissingKeys(t *testing.T) {
	t.Setenv("FROM_ENV", "x")
	got := missingKeys([]string{"SAVED", "FROM_ENV", "ABSENT"}, map[string]string{"SAVED": "v"})
	if len(got) != 1 || got[0] != "ABSENT" {
		t.Errorf("missingKeys = %v", got)
	}
}

func TestRunCmd_DoesNotParseAgentFlags(t *testing.T) {
	if err := runCmd.Flags().Parse([]string{"foo", "--dry-run", "-v"}); err != nil {
		t.Fatalf("flags after the agent name must be passed through: %v", err)
	}
	if got := runCmd.Flags().Args(); strings.Join(got, " ") != "foo --dry-run -v" {
		t.Errorf("args = %v", got)
	}
}

func TestResolveInterpreter(t *testing.T) {
	prev := lookPath
	t.Cleanup(func() { lookPath = prev })

	onPath := map[string]bool{}
	lookPath = func(name string) (string, error) {
		if onPath[name] {
			return "/usr/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}

	in := "python '/a b/agent.py' --x"

	onPath["python3"] = true
	if got := resolveInterpreter(in); got != "python3 '/a b/agent.py' --x" {
		t.Errorf("fallback not applied: %q", got)
	}

	onPath["python"] = true
	if got := resolveInterpreter(in); got != in {
		t.Errorf("python available, entrypoint must be unchanged: %q", got)
	}

	if got := resolveInterpreter("node index.js"); got != "node index.js" {
		t.Errorf("unrelated interpreter changed: %q", got)
	}
}
