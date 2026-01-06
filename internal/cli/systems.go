package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of systems/consoles",
	Long:  "Retrieves the complete list of systems (consoles) available in Screenscraper",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSystemsList()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(systemsCmd)
}
