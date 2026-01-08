package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Get list of classifications",
	Long:  "Retrieves the list of all game classifications/ratings (ESRB, PEGI, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetClassificationsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Classifications, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderClassificationsList(resp.Response.Classifications, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(classificationsCmd)
}
