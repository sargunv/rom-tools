package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Get list of classifications",
	Long:  "Retrieves the list of all game classifications/ratings (ESRB, PEGI, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetClassificationsList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Classifications, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderClassificationsList(resp.Response.Classifications, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(classificationsCmd)
}
