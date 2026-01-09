package detail

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/clients/screenscraper"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"

	"github.com/spf13/cobra"
)

var (
	systemDetailID int
)

var systemCmd = &cobra.Command{
	Use:     "system",
	Short:   "Get detailed system information",
	Long:    "Retrieves detailed information about a specific system/console",
	Example: `  rom-tools screenscraper detail system --id=1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetSystemsList()
		if err != nil {
			return err
		}

		// Find system by ID
		var found *screenscraper.System
		for i := range resp.Response.Systems {
			if resp.Response.Systems[i].ID == systemDetailID {
				found = &resp.Response.Systems[i]
				break
			}
		}

		if found == nil {
			return fmt.Errorf("system with ID %d not found", systemDetailID)
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(found, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderSystemDetail(*found, lang))
		return nil
	},
}

func init() {
	systemCmd.Flags().IntVar(&systemDetailID, "id", 0, "System ID")
	systemCmd.MarkFlagRequired("id")
	Cmd.AddCommand(systemCmd)
}
