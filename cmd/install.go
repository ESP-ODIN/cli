package cmd

import (
	"fmt"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [nom]",
	Short: "Installe un agent depuis le registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]

		ui.Blank()
		fmt.Printf("  %s %s\n",
			ui.StyleHeader.Render("Installation de"),
			ui.StyleAccent.Render(agentName),
		)
		ui.Blank()

		var agent internal.Agent

		err := ui.RunWithSpinner("Récupération depuis le registry...", func() error {
			var err error
			agent, err = internal.FetchAgentByName(agentName)
			return err
		})

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		ui.Step("Résolution du manifest...")
		ui.Step("Vérification de l'agent...")

		reg := internal.LoadRegistry()
		reg.Agents[agentName] = agent

		if err := internal.SaveRegistry(reg); err != nil {
			ui.PrintError("Erreur de sauvegarde : " + err.Error())
			ui.Blank()
			return
		}

		ui.Blank()
		ui.PrintSuccess(fmt.Sprintf("%s@%s installé", agent.Name, agent.Version))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
