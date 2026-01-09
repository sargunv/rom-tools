package screenscraper

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/clients/screenscraper"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

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

		resp, err := shared.Client.SearchGame(params)
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Games, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderGamesList(resp.Response.Games, lang))
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchSystemID, "system", "s", "", "System ID to filter results")
	Cmd.AddCommand(searchCmd)
}
