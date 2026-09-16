package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [nom]",
	Short: "Exécute un agent installé",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]

		reg := internal.LoadRegistry()
		agent, exists := reg.Agents[agentName]
		if !exists {
			ui.Blank()
			ui.PrintError(fmt.Sprintf("Agent '%s' not installed.", agentName))
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Run: odin install %s", agentName)))
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

		var command *exec.Cmd
		if runtime.GOOS == "windows" {
			command = exec.Command("cmd", "/C", agent.Run)
		} else {
			command = exec.Command("sh", "-c", agent.Run)
		}

		output, err := command.Output()
		if err != nil {
			ui.PrintError("Execution failed: " + err.Error())
			ui.Blank()
			return
		}

		fmt.Println(ui.StyleDesc.Render("  " + string(output)))
		ui.PrintSuccess("Done")
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
