package cmd

import (
	"fmt"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Affiche les agents disponibles dans le registry",
	Run: func(cmd *cobra.Command, args []string) {
		var agents []internal.Agent

		err := ui.RunWithSpinner("Récupération du registry...", func() error {
			var err error
			agents, err = internal.FetchAgentList()
			return err
		})

		ui.Blank()

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		fmt.Println(ui.StyleHeader.Render("  Agents disponibles"))
		ui.PrintDivider()
		ui.Blank()

		for _, agent := range agents {
			fmt.Printf("  %s  %s  %s\n",
				ui.StyleName.Render(agent.Name),
				ui.StyleVersion.Render("v"+agent.Version),
				ui.StyleMuted.Render("— "+agent.Description),
			)
		}

		ui.Blank()
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) disponible(s)", len(agents))))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
