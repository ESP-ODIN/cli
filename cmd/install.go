package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ESP-ODIN/cli/internal"
	"github.com/ESP-ODIN/cli/ui"
	"github.com/spf13/cobra"
)

var installRepo string

var installCmd = &cobra.Command{
	Use:   "install [nom[@version]]",
	Short: "Installe un agent depuis le registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		name, version := parseNameVersion(input)

		ui.Blank()
		fmt.Printf("  %s %s\n",
			ui.StyleHeader.Render("Installation de"),
			ui.StyleAccent.Render(input),
		)
		ui.Blank()

		var agent internal.Agent

		err := ui.RunWithSpinner("Récupération depuis le registry...", func() error {
			var err error
			if installRepo != "" {
				agent, err = internal.FetchManifestFromGitHub(installRepo)
			} else {
				agent, err = internal.FetchPackage(name, version)
			}
			return err
		})

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		reg := internal.LoadRegistry()
		reg.Agents[agent.Name] = agent

		if err := internal.SaveRegistry(reg); err != nil {
			ui.PrintError("Erreur de sauvegarde : " + err.Error())
			ui.Blank()
			return
		}

		// Clone le repo de l'agent si un Repo est défini (agents GitHub)
		if agent.Repo != "" {
			agentDir := internal.AgentDir(agent.Name)
			err = ui.RunWithSpinner("Cloning agent repository...", func() error {
				os.RemoveAll(agentDir)
				cmd := exec.Command("git", "clone", "--depth=1",
					"https://github.com/"+agent.Repo+".git", agentDir)
				cmd.Stdout = nil
				cmd.Stderr = nil
				return cmd.Run()
			})
			if err != nil {
				ui.PrintError("Erreur de clonage : " + err.Error())
				ui.Blank()
				return
			}

			reqFile := agentDir + "/requirements.txt"
			if _, statErr := os.Stat(reqFile); statErr == nil {
				_ = ui.RunWithSpinner("Installing dependencies...", func() error {
					cmd := exec.Command("pip3", "install", "-r", reqFile, "--quiet")
					return cmd.Run()
				})
			}
		}

		if len(agent.Requires) > 0 {
			keys := internal.LoadKeys()
			needsSave := false

			ui.Blank()
			fmt.Println(ui.StyleMuted.Render("  This agent requires the following API keys:"))
			ui.Blank()

			reader := bufio.NewReader(os.Stdin)
			for _, keyName := range agent.Requires {
				if _, already := keys[keyName]; already {
					fmt.Printf("  %s %s\n", ui.StyleMuted.Render(keyName+":"), ui.StyleMuted.Render("(already set)"))
					continue
				}
				fmt.Printf("  %s ", ui.StyleAccent.Render(keyName+":"))
				value, _ := reader.ReadString('\n')
				value = strings.TrimSpace(value)
				if value != "" {
					keys[keyName] = value
					needsSave = true
				}
			}

			if needsSave {
				if err := internal.SaveKeys(keys); err != nil {
					ui.PrintError("Erreur de sauvegarde des clés : " + err.Error())
					ui.Blank()
					return
				}
			}
			ui.Blank()
		}

		ui.PrintSuccess(fmt.Sprintf("%s@%s installé", agent.Name, agent.Version))
		ui.Blank()
	},
}

// parseNameVersion splits "name@version" into (name, version).
// If no version is specified, returns "latest".
func parseNameVersion(input string) (string, string) {
	parts := strings.SplitN(input, "@", 2)
	if len(parts) == 2 && parts[1] != "" {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}

func init() {
	installCmd.Flags().StringVar(&installRepo, "repo", "", "GitHub owner/repo à utiliser directement (bypass registry)")
	rootCmd.AddCommand(installCmd)
}
