package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ESP-ODIN/cli/odin-cli/internal"
	"github.com/ESP-ODIN/cli/odin-cli/ui"
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
			ui.StyleHeader.Render("Installing"),
			ui.StyleAccent.Render(agentName),
		)
		ui.Blank()

		var agent internal.Agent

		err := ui.RunWithSpinner("Fetching from registry...", func() error {
			url := fmt.Sprintf("http://localhost:3500/agents/%s", agentName)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("impossible de contacter le registry : %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				return fmt.Errorf("agent '%s' introuvable", agentName)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("erreur de lecture : %w", err)
			}

			return json.Unmarshal(body, &agent)
		})

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		ui.Step("Resolving manifest...")
		ui.Step("Verifying agent...")

		reg := internal.LoadRegistry()
		reg.Agents[agentName] = agent

		if err := internal.SaveRegistry(reg); err != nil {
			ui.PrintError("Erreur de sauvegarde : " + err.Error())
			ui.Blank()
			return
		}

		ui.Blank()
		ui.PrintSuccess(fmt.Sprintf("%s@%s installed", agent.Name, agent.Version))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
