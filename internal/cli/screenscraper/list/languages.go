package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var languagesCmd = &cobra.Command{
	Use:   "languages",
	Short: "Get list of languages",
	Long:  "Retrieves the list of all languages",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetLanguagesList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Languages, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderLanguagesList(resp.Response.Languages, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(languagesCmd)
}
