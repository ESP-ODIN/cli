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
	Long: "Recherche des agents dans le registry ODIN par nom ou description.\n" +
		"Sans mot-clé, liste tous les agents disponibles.",
	Example: "  odin search\n  odin search changelog",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = strings.TrimSpace(args[0])
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
			if query != "" {
				fmt.Println(ui.StyleMuted.Render("  Lancez : odin search  (pour tout lister)"))
			}
			ui.Blank()
			return
		}

		installed := internal.LoadRegistry().Agents

		for _, agent := range agents {
			printSearchResult(agent, installed)
		}

		fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("  %d agent(s) trouvé(s)", len(agents))))
		fmt.Println(ui.StyleMuted.Render("  Installer : odin install <nom>"))
		ui.Blank()
	},
}

func printSearchResult(agent internal.Agent, installed map[string]internal.Agent) {
	line := "  " + ui.StyleName.Render(agent.Name)
	if agent.IsOfficialPick {
		line += "  " + ui.StyleAccent.Render("★ officiel")
	}
	if local, ok := installed[agent.Name]; ok {
		tag := "installé"
		if local.Version != "" {
			tag += " v" + local.Version
		}
		line += "  " + ui.StyleSuccess.Render("✓ "+tag)
	}
	fmt.Println(line)

	if agent.Description != "" {
		fmt.Println("    " + ui.StyleDesc.Render(agent.Description))
	}

	meta := joinNonEmpty(" · ",
		agent.Category,
		agent.Runtime,
		agent.AgentType,
		fmt.Sprintf("%d téléchargement(s)", agent.DownloadsCount),
	)
	fmt.Println("    " + ui.StyleMuted.Render(meta))
	ui.Blank()
}

// joinNonEmpty joins the non-empty parts with sep.
func joinNonEmpty(sep string, parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
