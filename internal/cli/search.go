package cli

import (
	"encoding/json"
	"fmt"

	"sargunv/screenscraper-go/client"

	"github.com/spf13/cobra"
)

var (
	searchSystemID string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for games by name",
	Long:  "Searches for games by name, returns up to 30 results sorted by probability",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := screenscraper.SearchGameParams{
			Query:    args[0],
			SystemID: searchSystemID,
		}

		resp, err := client.SearchGame(params)
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
	searchCmd.Flags().StringVarP(&searchSystemID, "system", "s", "", "System ID to filter results")
	rootCmd.AddCommand(searchCmd)
}
