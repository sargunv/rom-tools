package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Get infrastructure/server information",
	Long:  "Retrieves Screenscraper infrastructure info including CPU usage, API load, and status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetInfraInfo()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp.Response.Servers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infraCmd)
}
