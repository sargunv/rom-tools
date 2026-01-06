package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var gameInfoTypesCmd = &cobra.Command{
	Use:   "game-info-types",
	Short: "Get list of game info types",
	Long:  "Retrieves the list of all available info types for games",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetGameInfoTypesList()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp.Response.InfoTypes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	listCmd.AddCommand(gameInfoTypesCmd)
}
