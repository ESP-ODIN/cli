package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Affiche les agents disponibles dans le registry",
	Run: func(cmd *cobra.Command, args []string) {
		var agents []internal.Agent

		err := ui.RunWithSpinner("Fetching registry...", func() error {
			resp, err := http.Get("http://localhost:3500/agents")
			if err != nil {
				return fmt.Errorf("impossible de contacter le registry : %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("erreur de lecture : %w", err)
			}

			return json.Unmarshal(body, &agents)
		})

		ui.Blank()

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		fmt.Println(ui.StyleHeader.Render("  Available agents"))
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
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) in registry", len(agents))))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
