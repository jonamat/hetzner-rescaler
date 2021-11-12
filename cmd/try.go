package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/jonamat/hetzner-rescaler/pkg/config"
	"github.com/jonamat/hetzner-rescaler/pkg/rescaler"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(tryCmd)
	tryCmd.Flags().BoolP("skip", "s", false, "Skip all user interactions")
}

/* Try command */
var tryCmd = &cobra.Command{
	Use:   "try",
	Short: "Try a complete rescale cycle",
	Long:  "Try a complete rescale cycle",
	Run:   RunTry,
}

/* Run fn for try command */
func RunTry(cmd *cobra.Command, args []string) {
	skip, err := cmd.Flags().GetBool("skip")
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	// Get the configuration from viper
	if err := config.CheckEnvs(); err != nil {
		color.Red("Error: %s", err.Error())
		cmd.Help()
		return
	}
	hCloudToken := viper.GetString("HCLOUD_TOKEN")
	serverId := viper.GetInt("SERVER_ID")
	topServerName := viper.GetString("TOP_SERVER_NAME")
	baseServerName := viper.GetString("BASE_SERVER_NAME")

	// Create hetzner Cloud API client
	client := hcloud.NewClient(hcloud.WithToken(hCloudToken))

	// Get server
	server, _, err := client.Server.GetByID(context.Background(), serverId)
	if err != nil {
		color.Red("Error while getting server: ", err.Error())
		return
	}
	if server == nil {
		color.Red("Server not found")
		return
	}

	// Print info about current configuration
	fmt.Printf(`The server named "%s" with ID %s, currently of type %s, will be:
→ Upgraded to server type %s
→ Downgraded to server type %s `,
		color.GreenString(server.Name),
		color.GreenString(strconv.Itoa(server.ID)),
		color.GreenString(server.ServerType.Name),
		color.GreenString(topServerName),
		color.GreenString(baseServerName),
	)

	// Ask for confirmation if --skip is not set
	if !skip {
		confirmInput := promptui.Prompt{
			Label: "Do you want to try this configuration? (y/n)",
		}
		confirm, err := confirmInput.Run()
		if err != nil {
			color.Red("Error: %s", err.Error())
			return
		}
		fmt.Printf("\n\n")

		if confirm != "y" {
			color.Red("Operation aborted")
			return
		}
	}

	/* --------------------------------- Rescale -------------------------------- */
	color.Green("Start upgrading server...")

	if err := rescaler.Rescale(client, server, topServerName); err != nil {
		color.Red("Error while resizing server: ", err.Error())
		return
	}

	// Update the server instance
	server, _, err = client.Server.GetByID(context.Background(), serverId)
	if err != nil {
		fmt.Println(color.RedString("Error while getting server: ", err.Error()))
		return
	}
	if server == nil {
		fmt.Println(color.RedString("Error: Server not found"))
		return
	}

	color.Green("Server successfully upgraded\n\n")
	color.Green("Start downgrading server...")

	if err := rescaler.Rescale(client, server, baseServerName); err != nil {
		color.Red("Error while resizing server: ", err.Error())
		return
	}
	color.Green("Server successfully downgraded\n\n")

	color.New(color.FgGreen).Add(color.Bold).Println("The rescale cycle has been completed succefully")
}
