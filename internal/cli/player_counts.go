package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var playerCountsCmd = &cobra.Command{
	Use:   "player-counts",
	Short: "Get list of player counts",
	Long:  "Retrieves the list of all player counts (1 player, 2 players, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetPlayerCountsList()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp.Response.PlayerCounts, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	listCmd.AddCommand(playerCountsCmd)
}
