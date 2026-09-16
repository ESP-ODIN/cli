package cmd

import (
	"fmt"
	"os"

	"github.com/ESP-ODIN/cli/odin-cli/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "odin",
	Short: "Odin — package manager d'agents IA",
	Long: ui.StyleMuted.Render("  ᚩ  ") + ui.StyleAccent.Render("ODIN") + "\n" + ui.StyleMuted.Render(ui.Tagline),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
