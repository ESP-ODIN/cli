package cmd

import (
	"bytes"
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
	Use:   "run [nom]",
	Short: "Exécute un agent installé",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]
		extraArgs := args[1:]

		reg := internal.LoadRegistry()
		agent, exists := reg.Agents[agentName]
		if !exists {
			ui.Blank()
			ui.PrintError(fmt.Sprintf("Agent '%s' non installé.", agentName))
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Lancez : odin install %s", agentName)))
			ui.Blank()
			return
		}

		ui.Blank()
		fmt.Printf("  %s %s\n",
			ui.StyleHeader.Render("Running"),
			ui.StyleAccent.Render(agentName),
		)
		ui.PrintDivider()
		ui.Blank()

		agentDir := internal.AgentDir(agentName)

		// Rewrite relative script paths to absolute so the agent can be run
		// from the user's cwd (which is the git repo the agent will operate on).
		runCommand := makeAbsolute(agent.Run, agentDir)
		if len(extraArgs) > 0 {
			runCommand = runCommand + " " + strings.Join(extraArgs, " ")
		}

		var command *exec.Cmd
		if runtime.GOOS == "windows" {
			command = exec.Command("cmd", "/C", runCommand)
		} else {
			command = exec.Command("sh", "-c", runCommand)
		}

		// Execute from the user's working directory so git log reads their repo.
		cwd, _ := os.Getwd()
		command.Dir = cwd

		keys := internal.LoadKeys()
		command.Env = os.Environ()
		for k, v := range keys {
			command.Env = append(command.Env, k+"="+v)
		}

		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr

		if err := command.Run(); err != nil {
			if stderr.Len() > 0 {
				fmt.Println(ui.StyleMuted.Render(stderr.String()))
			}
			ui.PrintError("Execution failed: " + err.Error())
			ui.Blank()
			return
		}

		fmt.Println(ui.StyleDesc.Render("  " + stdout.String()))
		ui.PrintSuccess("Done")
		ui.Blank()
	},
}

// makeAbsolute rewrites the first script argument in a shell command to its
// absolute path inside agentDir, e.g. "python3 agent.py" → "python3 /home/.../.odin/agents/foo/agent.py"
func makeAbsolute(command, agentDir string) string {
	parts := strings.SplitN(command, " ", 2)
	if len(parts) < 2 {
		return command
	}
	interpreter := parts[0]
	rest := parts[1]

	// Split rest into script + remaining args
	restParts := strings.SplitN(rest, " ", 2)
	script := restParts[0]

	if !filepath.IsAbs(script) {
		script = filepath.Join(agentDir, script)
	}

	if len(restParts) > 1 {
		return interpreter + " " + script + " " + restParts[1]
	}
	return interpreter + " " + script
}

func init() {
	rootCmd.AddCommand(runCmd)
}
