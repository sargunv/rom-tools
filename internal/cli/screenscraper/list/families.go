package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var familiesCmd = &cobra.Command{
	Use:   "families",
	Short: "Get list of families",
	Long:  "Retrieves the list of all game families (e.g., Sonic, Mario, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetFamiliesList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Families, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderFamiliesList(resp.Response.Families, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(familiesCmd)
}
