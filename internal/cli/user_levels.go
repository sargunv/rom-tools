package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var userLevelsCmd = &cobra.Command{
	Use:   "user-levels",
	Short: "Get list of user levels",
	Long:  "Retrieves the list of all ScreenScraper user levels",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetUserLevelsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.UserLevels, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderUserLevelsList(resp.Response.UserLevels, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(userLevelsCmd)
}
