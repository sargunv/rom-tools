package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var supportTypesCmd = &cobra.Command{
	Use:   "support-types",
	Short: "Get list of support types",
	Long:  "Retrieves the list of all support types (cartridge, CD, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSupportTypesList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.SupportTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderSupportTypesList(resp.Response.SupportTypes, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(supportTypesCmd)
}
