package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "Get list of regions",
	Long:  "Retrieves the list of all regions (USA, Europe, Japan, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetRegionsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Regions, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderRegionsList(resp.Response.Regions, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(regionsCmd)
}
