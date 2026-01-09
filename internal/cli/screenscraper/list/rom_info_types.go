package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var romInfoTypesCmd = &cobra.Command{
	Use:   "rom-info-types",
	Short: "Get list of ROM info types",
	Long:  "Retrieves the list of all available info types for ROMs",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetROMInfoTypesList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.InfoTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderROMInfoTypesList(resp.Response.InfoTypes, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(romInfoTypesCmd)
}
