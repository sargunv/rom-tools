package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var supportTypesCmd = &cobra.Command{
	Use:   "support-types",
	Short: "Get list of support types",
	Long:  "Retrieves the list of all support types (cartridge, CD, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetSupportTypesList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.SupportTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderSupportTypesList(resp.Response.SupportTypes, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(supportTypesCmd)
}
