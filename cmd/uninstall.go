package cmd

import (
	"fmt"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall [nom]",
	Short:   "Désinstalle un agent installé localement",
	Aliases: []string{"remove", "rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]

		ui.Blank()

		reg := internal.LoadRegistry()
		agent, exists := reg.Agents[agentName]
		if !exists {
			ui.PrintError(fmt.Sprintf("Agent '%s' non installé.", agentName))
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  Lancez : odin install %s", agentName)))
			ui.Blank()
			return
		}

		delete(reg.Agents, agentName)

		if err := internal.SaveRegistry(reg); err != nil {
			ui.PrintError("Erreur de sauvegarde : " + err.Error())
			ui.Blank()
			return
		}

		ui.PrintSuccess(fmt.Sprintf("%s@%s désinstallé", agent.Name, agent.Version))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
