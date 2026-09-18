package cmd

import (
	"fmt"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the odin CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Blank()
		fmt.Printf("  %s %s\n",
			ui.StyleAccent.Render("odin"),
			ui.StyleMuted.Render("v"+internal.Version),
		)
		ui.Blank()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
