package cmd

import (
	"fmt"
	"sort"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Affiche les agents installés",
	Long: "Affiche les agents installés localement (~/.odin/registry.json).\n" +
		"Pour voir les agents disponibles dans le registry : odin search",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reg := internal.LoadRegistry()

		ui.Blank()
		fmt.Println(ui.StyleHeader.Render("  Agents installés"))
		ui.PrintDivider()
		ui.Blank()

		if len(reg.Agents) == 0 {
			fmt.Println(ui.StyleMuted.Render("  Aucun agent installé."))
			fmt.Println(ui.StyleMuted.Render("  Lancez : odin search  puis  odin install <nom>"))
			ui.Blank()
			return
		}

		names := make([]string, 0, len(reg.Agents))
		for name := range reg.Agents {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			agent := reg.Agents[name]

			line := "  " + ui.StyleName.Render(name)
			if agent.Version != "" {
				line += "  " + ui.StyleVersion.Render("v"+agent.Version)
			}
			if agent.Description != "" {
				line += "  " + ui.StyleMuted.Render("— "+agent.Description)
			}
			fmt.Println(line)
		}

		ui.Blank()
		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) installé(s)", len(reg.Agents))))
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
