package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var genresCmd = &cobra.Command{
	Use:   "genres",
	Short: "Get list of genres",
	Long:  "Retrieves the list of all game genres",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetGenresList()
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
	rootCmd.AddCommand(genresCmd)
}
