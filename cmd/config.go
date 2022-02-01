package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/spaceweasel/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/* ServerType extended type */
type extServerType struct {
	*hcloud.ServerType
	Type       string
	Memory     int
	PriceHour  string
	PriceMonth string
}

var weekDays = []string{
	time.Monday.String(),
	time.Tuesday.String(),
	time.Wednesday.String(),
	time.Thursday.String(),
	time.Friday.String(),
	time.Saturday.String(),
	time.Sunday.String(),
}

/* Config command */
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Create the configuration file",
	Long:  "Assistant for creating the configuration file.\nIf --config flag is not provided, the config file will be saved in ~/.hetzner-rescaler.yaml",
	Run:   RunConfig,
}

/* Templates for server descriptions */
var templates = &promptui.SelectTemplates{
	Label:    "{{ . }}",
	Active:   fmt.Sprintf("{{ \"%s\" | white }} {{ .Name | underline }}", promptui.IconSelect),
	Inactive: "{{ .Name }}",
	Selected: fmt.Sprintf("{{ \"%s\" | green }} {{ .Name | green }} | vCPU: {{ .Cores }} {{ .Type }} | Memory: {{ .Memory }}GB | Disk size: {{ .Disk }}GB | Price: {{ .PriceHour }}/h {{ .PriceMonth }}/month", promptui.IconGood),
	Details: `
{{ "vCPU:" | green }} {{ .Cores }}
{{ "CPU type:" | green }} {{ .Type }}
{{ "Memory:" | green }} {{ .Memory }} GB
{{ "Disk:" | green }} {{ .Disk }} GB
{{ "Price:" | green }} {{ .PriceHour }}/h | {{ .PriceMonth }}/month`,
}

/* 24h format comma separated time validator */
func validateTimeFormat(s string) error {
	_, err := time.Parse("15:04", s)
	if err != nil {
		return fmt.Errorf("use a 24h format, like 20:30, or 3:08")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
}

/* Run fn for config command */
func RunConfig(cmd *cobra.Command, args []string) {
	/* ---------------------------------- Token --------------------------------- */
	color.Yellow("### HCLOUD TOKEN")
	tokenInput := promptui.Prompt{
		Label: "Enter your Hetzner Cloud token",
	}
	token, _ := tokenInput.Run()

	// Create hetzner Cloud API client
	client := hcloud.NewClient(hcloud.WithToken(token))
	fmt.Printf("\n\n")

	/* ------------------------------ Server select ----------------------------- */
	color.Yellow("### SERVER SELECT")

	// List of all the servers
	servers, err := client.Server.All(context.Background())
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	// List of all server names
	var serverNames []string
	for _, server := range servers {
		serverNames = append(serverNames, server.Name)
	}

	// Prompt server selection
	serverSelect := promptui.Select{
		Label: "Select the server you want to rescale",
		Items: serverNames,
	}

	index, _, err := serverSelect.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	// The selected server to rescale
	server := servers[index]

	// Show server current status
	fmt.Print("\n\n")
	fmt.Printf("Selected server \"%s\" with ID %s\n",
		color.GreenString(server.Name),
		color.GreenString(strconv.Itoa(server.ID)),
	)
	fmt.Printf("This server is currently of type %s with a disk size of %sGB\n\n",
		color.GreenString(server.ServerType.Name),
		color.GreenString(strconv.Itoa(server.PrimaryDiskSize)),
	)

	/* ---------------------------- Get server types ---------------------------- */
	// List of all server types
	serverTypes, err := client.ServerType.All(context.Background())
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	// List of all server types elegible for the selected server
	var elegibleServerTypes []extServerType
	for _, s := range serverTypes {
		if s.Disk >= server.PrimaryDiskSize {
			var se = extServerType{
				ServerType: s,
				Memory:     int(s.Memory),
				PriceHour:  s.Pricings[0].Hourly.Gross[:7] + s.Pricings[0].Hourly.Currency,
				PriceMonth: s.Pricings[0].Monthly.Gross[:5] + s.Pricings[0].Monthly.Currency,
			}

			switch {
			case strings.Contains(s.Name, "cpx"):
				se.Type = "Shared AMD CPU"
			case strings.Contains(s.Name, "ccx"):
				se.Type = "Dedicaded AMD or Intel CPU"
			case strings.Contains(s.Name, "cx"):
				se.Type = "Shared Intel CPU"
			}

			elegibleServerTypes = append(elegibleServerTypes, se)
		}
	}

	/* ---------------------------- Base server type ---------------------------- */
	color.Yellow("\n\n### BASE SERVER TYPE")

	// Prompt server type selection
	baseServerTypeSelect := promptui.Select{
		Label:     "What type of base (cheaper) server type you want to rescale to?",
		Items:     elegibleServerTypes,
		Templates: templates,
	}

	index, _, err = baseServerTypeSelect.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	baseServerType := elegibleServerTypes[index]

	/* ----------------------------- Top server type ---------------------------- */
	color.Yellow("\n\n### TOP SERVER TYPE")

	// Prompt server type selection
	topServerTypeSelect := promptui.Select{
		Label:     "What type of top server type you want to rescale to?",
		Items:     elegibleServerTypes,
		Templates: templates,
	}

	index, _, err = topServerTypeSelect.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	topServerType := elegibleServerTypes[index]

	/* --------------------------------- Checks --------------------------------- */
	if topServerType.ID == baseServerType.ID {
		color.Red("The top server type must be different from the base server type")
		return
	}

	/* ------------------------------- Start time ------------------------------- */
	color.Yellow("\n\n### TOP SERVER START TIME")

	startTimeInput := promptui.Prompt{
		Label:    "When should the server upgrade to the top type? (local time, 24h format)",
		Validate: validateTimeFormat,
		Default:  "09:00",
	}

	hourStart, err := startTimeInput.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	/* -------------------------------- Stop time ------------------------------- */
	color.Yellow("\n\n### TOP SERVER STOP TIME")

	stopTimeInput := promptui.Prompt{
		Label:    "When should the server downgrade to the base type? (local time, 24h format)",
		Validate: validateTimeFormat,
		Default:  "20:00",
	}

	hourStop, err := stopTimeInput.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	/* -------------------------------- Week days ------------------------------- */
	color.Yellow("\n\n### RESCALE WEEK DAYS")

	weekDaysInput := promptui.MultiSelect{
		Label: "Select the week days you want to rescale your server (space for select, enter for confirm)",
		Items: weekDays,
	}

	weekDaysIndexes, err := weekDaysInput.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	selectedWeekDays := []string{}
	for _, day := range weekDaysIndexes {
		selectedWeekDays = append(selectedWeekDays, weekDays[day])
	}

	/* --------------------------------- Summary -------------------------------- */
	color.Yellow("\n\n### SUMMARY")

	fmt.Printf(`The server named "%s" with ID %s, currently of type %s, will be:
→ Upgraded to server type %s every selected day at %s
→ Downgraded to server type %s every selected day at %s

The days of the week selected for rescaling are: %s`,
		color.GreenString(server.Name),
		color.GreenString(strconv.Itoa(server.ID)),
		color.GreenString(server.ServerType.Name),
		color.GreenString(topServerType.Name),
		color.GreenString(hourStart),
		color.GreenString(baseServerType.Name),
		color.GreenString(hourStop),
		color.GreenString(strings.Join(selectedWeekDays, ", ")),
	)

	/* --------------------------------- Confirm -------------------------------- */
	fmt.Print("\n\n")
	confirmInput := promptui.Prompt{
		Label: "Is this configuration correct? This operation will override previous configuration file (y/n)",
	}
	confirm, err := confirmInput.Run()
	if err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	if confirm != "y" {
		color.New(color.FgRed).Add(color.Bold).Println("Operation aborted")
		return
	}

	/* ------------------------------- Write config ------------------------------ */
	viper.Set("HCLOUD_TOKEN", token)
	viper.Set("SERVER_ID", server.ID)
	viper.Set("BASE_SERVER_NAME", baseServerType.Name)
	viper.Set("TOP_SERVER_NAME", topServerType.Name)
	viper.Set("HOUR_START", hourStart)
	viper.Set("HOUR_STOP", hourStop)
	viper.Set("WEEK_DAYS", selectedWeekDays)

	// Remove old config file if exists or create a new one
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = os.ExpandEnv("$HOME/.hetzner-rescaler.yaml")
	}

	if _, err := os.Stat(configPath); err != nil {
		if _, err := os.Create(configPath); err != nil {
			color.Red("Error: %s", err.Error())
			return
		}
	} else {
		if err := os.Remove(viper.ConfigFileUsed()); err != nil {
			color.Red("Error: %s", err.Error())
			return
		}
	}

	// Save the configuration
	if err := viper.WriteConfig(); err != nil {
		color.Red("Error: %s", err.Error())
		return
	}

	color.New(color.FgGreen).Add(color.Bold).Println("\n\nConfiguration saved\n")
}
