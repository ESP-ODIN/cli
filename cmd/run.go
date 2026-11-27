package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <nom> [args...]",
	Short: "Exécute un agent installé",
	Long: "Exécute un agent installé depuis le répertoire courant.\n" +
		"Les arguments après le nom sont transmis tels quels à l'agent,\n" +
		"et les clés API enregistrées sont injectées en variables d'environnement.",
	Example: "  odin run changelog-agent\n  odin run commit-helper --dry-run",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(runAgent(args[0], args[1:]))
	},
}

// runAgent executes an installed agent and returns the exit code to use.
func runAgent(agentName string, extraArgs []string) int {
	reg := internal.LoadRegistry()
	agent, exists := reg.Agents[agentName]
	if !exists {
		ui.Blank()
		ui.PrintError(fmt.Sprintf("Agent '%s' non installé.", agentName))
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Lancez : odin install %s", agentName)))
		ui.Blank()
		return 1
	}

	if strings.TrimSpace(agent.Run) == "" {
		ui.Blank()
		ui.PrintError(fmt.Sprintf("Aucun entrypoint défini pour '%s'.", agentName))
		fmt.Println(ui.StyleMuted.Render("  Le manifest odin.toml de l'agent doit déclarer un champ entrypoint."))
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Réinstallez l'agent : odin install %s", agentName)))
		ui.Blank()
		return 1
	}

	agentDir := internal.AgentDir(agentName)
	if agent.Repo != "" {
		if _, err := os.Stat(agentDir); err != nil {
			ui.Blank()
			ui.PrintError(fmt.Sprintf("Fichiers de l'agent introuvables (%s).", agentDir))
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Réinstallez l'agent : odin install %s", agentName)))
			ui.Blank()
			return 1
		}
	}

	keys := internal.LoadKeys()
	if missing := missingKeys(agent.Requires, keys); len(missing) > 0 {
		ui.Blank()
		ui.PrintError("Clé(s) API manquante(s) : " + strings.Join(missing, ", "))
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Réinstallez l'agent pour les renseigner : odin install %s", agentName)))
		ui.Blank()
		return 1
	}

	command := buildCommand(resolveInterpreter(makeAbsolute(agent.Run, agentDir)), extraArgs)

	cwd, err := os.Getwd()
	if err == nil {
		command.Dir = cwd
	}

	command.Env = append(os.Environ(),
		"ODIN_AGENT_NAME="+agentName,
		"ODIN_AGENT_DIR="+agentDir,
	)
	for k, v := range keys {
		command.Env = append(command.Env, k+"="+v)
	}

	// Stream I/O so interactive and long-running agents behave like a
	// normal process.
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			ui.Blank()
			ui.PrintError(fmt.Sprintf("%s s'est terminé avec le code %d", agentName, exitErr.ExitCode()))
			return exitErr.ExitCode()
		}
		ui.Blank()
		ui.PrintError("Impossible de lancer l'agent : " + err.Error())
		return 1
	}
	return 0
}

// buildCommand wraps the entrypoint in the platform shell and forwards
// extraArgs without re-interpreting them (no word splitting / injection).
func buildCommand(entrypoint string, extraArgs []string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		line := entrypoint
		for _, a := range extraArgs {
			line += " " + windowsQuote(a)
		}
		return exec.Command("cmd", "/C", line)
	}
	// sh -c '<entrypoint> "$@"' odin arg1 arg2 ...
	shArgs := append([]string{"-c", entrypoint + ` "$@"`, "odin"}, extraArgs...)
	return exec.Command("sh", shArgs...)
}

// makeAbsolute rewrites the script referenced by the entrypoint to its
// absolute path inside agentDir, so the agent can run from the user's cwd:
//
//	"python3 agent.py" → "python3 '/home/.../.odin/agents/foo/agent.py'"
//	"./run.sh --x"     → "'/home/.../.odin/agents/foo/run.sh' --x"
//
// Only the first of the two leading tokens that names an existing file in
// agentDir is rewritten; interpreters such as "python3" are left untouched.
func makeAbsolute(command, agentDir string) string {
	fields := strings.Fields(command)
	for i := 0; i < len(fields) && i < 2; i++ {
		tok := fields[i]
		if tok == "" || strings.HasPrefix(tok, "-") || filepath.IsAbs(tok) {
			continue
		}
		candidate := filepath.Join(agentDir, tok)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			fields[i] = shellQuote(candidate)
			return strings.Join(fields, " ")
		}
	}
	return command
}

// interpreterFallbacks maps interpreters commonly written in manifests to an
// alternative used when the first one is not on PATH (e.g. macOS ships
// python3 but no python).
var interpreterFallbacks = map[string]string{
	"python": "python3",
	"pip":    "pip3",
}

// lookPath is overridable in tests.
var lookPath = exec.LookPath

// resolveInterpreter swaps the entrypoint's interpreter for its fallback when
// the declared one is missing and the fallback is available.
func resolveInterpreter(entrypoint string) string {
	fields := strings.Fields(entrypoint)
	if len(fields) == 0 {
		return entrypoint
	}
	fallback, ok := interpreterFallbacks[fields[0]]
	if !ok {
		return entrypoint
	}
	if _, err := lookPath(fields[0]); err == nil {
		return entrypoint
	}
	if _, err := lookPath(fallback); err != nil {
		return entrypoint
	}
	// Replace only the leading token so quoted paths later on stay intact.
	trimmed := strings.TrimLeft(entrypoint, " \t")
	return fallback + strings.TrimPrefix(trimmed, fields[0])
}

func missingKeys(required []string, keys map[string]string) []string {
	var missing []string
	for _, k := range required {
		if keys[k] == "" && os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	return missing
}

func shellQuote(s string) string {
	if runtime.GOOS == "windows" {
		return windowsQuote(s)
	}
	if s != "" && !strings.ContainsAny(s, " \t\n'\"\\$`!*?[]{}()<>|&;#~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func windowsQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\"") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func init() {
	// Everything after the agent name belongs to the agent, including flags:
	// `odin run foo --dry-run` must not be parsed by cobra.
	runCmd.Flags().SetInterspersed(false)
	rootCmd.AddCommand(runCmd)
}
