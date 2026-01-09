package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var playerCountsCmd = &cobra.Command{
	Use:   "player-counts",
	Short: "Get list of player counts",
	Long:  "Retrieves the list of all player counts (1 player, 2 players, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetPlayerCountsList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.PlayerCounts, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderPlayerCountsList(resp.Response.PlayerCounts, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(playerCountsCmd)
}
