package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var genresCmd = &cobra.Command{
	Use:   "genres",
	Short: "Get list of genres",
	Long:  "Retrieves the list of all game genres",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetGenresList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Genres, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderGenresList(resp.Response.Genres, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(genresCmd)
}
