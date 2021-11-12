package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/jonamat/hetzner-rescaler/pkg/config"
	"github.com/jonamat/hetzner-rescaler/pkg/rescaler"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/* Start command */
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start rescale timers",
	Long:  "Start rescale timers",
	Run:   RunStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("skip", "s", false, "Skip all user interactions")
}

/* Run fn for start command */
func RunStart(cmd *cobra.Command, args []string) {
	skip, err := cmd.Flags().GetBool("skip")
	if err != nil {
		log.Println(color.RedString("Error: %s", err.Error()))
		return
	}

	// Get the configuration from viper
	if err := config.CheckEnvs(); err != nil {
		log.Println(color.RedString("Error: %s", err.Error()))
		cmd.Help()
		return
	}
	hCloudToken := viper.GetString("HCLOUD_TOKEN")
	serverId := viper.GetInt("SERVER_ID")
	topServerName := viper.GetString("TOP_SERVER_NAME")
	baseServerName := viper.GetString("BASE_SERVER_NAME")
	hourStart := viper.GetString("HOUR_START")
	hourStop := viper.GetString("HOUR_STOP")

	// Create hetzner Cloud API client
	client := hcloud.NewClient(hcloud.WithToken(hCloudToken))

	// Get server
	server, _, err := client.Server.GetByID(context.Background(), serverId)
	if err != nil {
		log.Println(color.RedString("Error while getting server: ", err.Error()))
		return
	}
	if server == nil {
		log.Println(color.RedString("Error: Server not found"))
		return
	}

	// Print info about current configuration
	fmt.Printf(`The server named "%s" with ID %s, currently of type %s, will be:
→ Upgraded to server type %s everyday at %s
→ Downgraded to server type %s everyday at %s`,
		color.GreenString(server.Name),
		color.GreenString(strconv.Itoa(server.ID)),
		color.GreenString(server.ServerType.Name),
		color.GreenString(topServerName),
		color.GreenString(hourStart),
		color.GreenString(baseServerName),
		color.GreenString(hourStop),
	)

	// Ask for confirmation if --skip is not set
	if !skip {
		confirmInput := promptui.Prompt{
			Label: "Do you want to start with this configuration? (y/n)",
		}
		confirm, err := confirmInput.Run()
		if err != nil {
			log.Println(color.RedString("Error: %s", err.Error()))
			return
		}

		if confirm != "y" {
			fmt.Println(color.RedString("Operation aborted"))
			return
		}
	}

	/* ------------------------------- Start timer ------------------------------ */
	log.Println(color.GreenString("Timer started\n"))

	for {
		hour, min, _ := time.Now().Clock()
		currentHour := fmt.Sprintf("%02d:%02d", hour, min)

		if currentHour == hourStart {
			log.Println(color.GreenString("Start upgrading server..."))

			if err := rescaler.Rescale(client, server, topServerName); err != nil {
				log.Println(color.RedString("Error while resizing server: ", err.Error()))
				return
			}

			// Update the server instance
			server, _, err = client.Server.GetByID(context.Background(), serverId)
			if err != nil {
				log.Println(color.RedString("Error while getting server: ", err.Error()))
				return
			}
			if server == nil {
				log.Println(color.RedString("Error: Server not found"))
				return
			}

			log.Println(color.GreenString("Server successfully upgraded\n"))
		}

		if currentHour == hourStop {
			log.Println(color.GreenString("Start downgrading server..."))

			if err := rescaler.Rescale(client, server, baseServerName); err != nil {
				log.Println(color.RedString("Error while resizing server: ", err.Error()))
				return
			}

			// Update the server instance
			server, _, err = client.Server.GetByID(context.Background(), serverId)
			if err != nil {
				log.Println(color.RedString("Error while getting server: ", err.Error()))
				return
			}
			if server == nil {
				log.Println(color.RedString("Error: Server not found"))
				return
			}

			log.Println(color.GreenString("Server successfully downgraded\n"))
		}

		time.Sleep(time.Second * 60)
	}
}
