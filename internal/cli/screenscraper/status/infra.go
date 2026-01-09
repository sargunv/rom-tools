package status

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Get infrastructure/server information",
	Long:  "Retrieves Screenscraper infrastructure info including CPU usage, API load, and status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetInfraInfo()
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Servers, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderInfra(resp.Response.Servers, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(infraCmd)
}
