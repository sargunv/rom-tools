package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var romTypesCmd = &cobra.Command{
	Use:   "rom-types",
	Short: "Get list of ROM types",
	Long:  "Retrieves the list of all ROM types",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetROMTypesList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.ROMTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderROMTypesList(resp.Response.ROMTypes, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(romTypesCmd)
}
