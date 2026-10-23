package cmd

import (
	"fmt"
	"strings"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [mot-clé]",
	Short: "Recherche des agents disponibles dans le registry",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		var agents []internal.Agent

		label := "Récupération du registry..."
		if query != "" {
			label = fmt.Sprintf("Recherche de « %s »...", query)
		}

		err := ui.RunWithSpinner(label, func() error {
			var err error
			agents, err = internal.SearchPackages(query)
			return err
		})

		ui.Blank()

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		title := "Agents disponibles"
		if query != "" {
			title = fmt.Sprintf("Résultats pour « %s »", query)
		}
		fmt.Println(ui.StyleHeader.Render("  " + title))
		ui.PrintDivider()
		ui.Blank()

		if len(agents) == 0 {
			fmt.Println(ui.StyleMuted.Render("  Aucun agent trouvé."))
			ui.Blank()
			return
		}

		for _, agent := range agents {
			versionStr := ""
			if agent.Version != "" {
				versionStr = ui.StyleVersion.Render("v" + agent.Version)
			}
			meta := ui.StyleMuted.Render(strings.Join([]string{agent.Category, agent.Runtime}, " · "))
			desc := ""
			if agent.Description != "" {
				desc = ui.StyleMuted.Render("— " + agent.Description)
			}
			fmt.Printf("  %s  %s  %s  %s\n",
				ui.StyleName.Render(agent.Name),
				versionStr,
				meta,
				desc,
			)
		}

		ui.Blank()
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) trouvé(s)", len(agents))))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
