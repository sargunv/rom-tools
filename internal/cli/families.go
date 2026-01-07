package cli

import (
	"encoding/json"
	"fmt"

	screenscraper "sargunv/screenscraper-go/client"
	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var familiesCmd = &cobra.Command{
	Use:   "families",
	Short: "Get list of families",
	Long:  "Retrieves the list of all game families (e.g., Sonic, Mario, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetFamiliesList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Families, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderFamiliesList(resp.Response.Families, lang))
		return nil
	},
}

var (
	familyDetailID int
)

var familyDetailCmd = &cobra.Command{
	Use:     "family",
	Short:   "Get detailed family information",
	Long:    "Retrieves detailed information about a specific game family",
	Example: `  screenscraper detail family --id=1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetFamiliesList()
		if err != nil {
			return err
		}

		// Find family by ID
		var found *screenscraper.Family
		for k := range resp.Response.Families {
			family := resp.Response.Families[k]
			if family.ID == familyDetailID {
				found = &family
				break
			}
		}

		if found == nil {
			return fmt.Errorf("family with ID %d not found", familyDetailID)
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
		fmt.Print(format.RenderFamilyDetail(*found, lang))
		return nil
	},
}

func init() {
	familyDetailCmd.Flags().IntVar(&familyDetailID, "id", 0, "Family ID")
	familyDetailCmd.MarkFlagRequired("id")
	detailCmd.AddCommand(familyDetailCmd)
	listCmd.AddCommand(familiesCmd)
}
