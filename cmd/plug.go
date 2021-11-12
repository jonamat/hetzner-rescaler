package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

/* Plug command */
var plugCmd = &cobra.Command{
	Use:   "plug",
	Short: "Configure and start immediately",
	Long:  "Configure and start immediately\nShorthand for config & start --skip",
	Run:   runPlug,
}

func init() {
	rootCmd.AddCommand(plugCmd)
}

/* Run fn for plug command */
func runPlug(cmd *cobra.Command, args []string) {
	// Avoid panic if RunConfig fails
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error:", err)
		}
		os.Exit(1)
	}()
	RunConfig(cmd, args)

	// Set skip flag true manually
	cmd.Flags().BoolP("skip", "s", true, "")

	RunStart(cmd, args)
}
