package list

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of systems/consoles",
	Long:  "Retrieves the complete list of systems (consoles) available in Screenscraper",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetSystemsList()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Systems, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderSystemsList(resp.Response.Systems, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(systemsCmd)
}
