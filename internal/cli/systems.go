package cli

import (
	"encoding/json"
	"fmt"

	screenscraper "sargunv/screenscraper-go/client"
	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of systems/consoles",
	Long:  "Retrieves the complete list of systems (consoles) available in Screenscraper",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSystemsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Systems, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderSystemsList(resp.Response.Systems, lang))
		return nil
	},
}

var (
	systemDetailID int
)

var systemDetailCmd = &cobra.Command{
	Use:     "system",
	Short:   "Get detailed system information",
	Long:    "Retrieves detailed information about a specific system/console",
	Example: `  screenscraper detail system --id=1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSystemsList()
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

		if jsonOutput {
			formatted, err := json.MarshalIndent(found, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderSystemDetail(*found, lang))
		return nil
	},
}

func init() {
	systemDetailCmd.Flags().IntVar(&systemDetailID, "id", 0, "System ID")
	systemDetailCmd.MarkFlagRequired("id")
	detailCmd.AddCommand(systemDetailCmd)
	listCmd.AddCommand(systemsCmd)
}
