package cmd

import (
	"fmt"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Affiche les agents installés",
	Run: func(cmd *cobra.Command, args []string) {
		reg := internal.LoadRegistry()

		ui.Blank()
		fmt.Println(ui.StyleHeader.Render("  Installed agents"))
		ui.PrintDivider()
		ui.Blank()

		if len(reg.Agents) == 0 {
			fmt.Println(ui.StyleMuted.Render("  No agents installed."))
			fmt.Println(ui.StyleMuted.Render("  Run: odin install <name>"))
			ui.Blank()
			return
		}

		for name, agent := range reg.Agents {
			fmt.Printf("  %s  %s  %s\n",
				ui.StyleName.Render(name),
				ui.StyleVersion.Render("v"+agent.Version),
				ui.StyleMuted.Render("— "+agent.Description),
			)
		}

		ui.Blank()
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) installed", len(reg.Agents))))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
