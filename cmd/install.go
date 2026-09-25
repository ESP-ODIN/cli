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
	Use:   "install [nom]",
	Short: "Installe un agent depuis le registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]

		ui.Blank()
		fmt.Printf("  %s %s\n",
			ui.StyleHeader.Render("Installation de"),
			ui.StyleAccent.Render(agentName),
		)
		ui.Blank()

		var agent internal.Agent

		err := ui.RunWithSpinner("Récupération depuis le registry...", func() error {
			var err error
			if installRepo != "" {
				agent, err = internal.FetchManifestFromGitHub(installRepo)
			} else {
				agent, err = internal.FetchAgentByName(agentName)
			}
			return err
		})

		if err != nil {
			ui.PrintError(err.Error())
			ui.Blank()
			return
		}

		ui.Step("Résolution du manifest...")
		ui.Step("Vérification de l'agent...")

		reg := internal.LoadRegistry()
		reg.Agents[agentName] = agent

		if err := internal.SaveRegistry(reg); err != nil {
			ui.PrintError("Erreur de sauvegarde : " + err.Error())
			ui.Blank()
			return
		}

		// Clone the agent's repo into ~/.odin/agents/<name>/
		agentDir := internal.AgentDir(agentName)
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

		// Install Python dependencies if requirements.txt exists
		reqFile := agentDir + "/requirements.txt"
		if _, statErr := os.Stat(reqFile); statErr == nil {
			_ = ui.RunWithSpinner("Installing dependencies...", func() error {
				cmd := exec.Command("pip3", "install", "-r", reqFile, "--quiet")
				return cmd.Run()
			})
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

func init() {
	installCmd.Flags().StringVar(&installRepo, "repo", "", "GitHub owner/repo à utiliser directement (bypass registry)")
	rootCmd.AddCommand(installCmd)
}
