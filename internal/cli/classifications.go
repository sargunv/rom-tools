package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	screenscraper "sargunv/screenscraper-go/client"
	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Get list of classifications",
	Long:  "Retrieves the list of all game classifications/ratings (ESRB, PEGI, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetClassificationsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Classifications, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderClassificationsList(resp.Response.Classifications, lang))
		return nil
	},
}

var (
	classificationDetailID   int
	classificationDetailName string
)

var classificationDetailCmd = &cobra.Command{
	Use:   "classification",
	Short: "Get detailed classification information",
	Long:  "Retrieves detailed information about a specific classification/rating",
	Example: `  screenscraper detail classification --id=1
  screenscraper detail classification --name=PEGI`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetClassificationsList()
		if err != nil {
			return err
		}

		// Find classification by ID or name
		var found *screenscraper.Classification
		for k, classification := range resp.Response.Classifications {
			if classificationDetailID != 0 && classification.ID == classificationDetailID {
				found = &classification
				break
			}
			if classificationDetailName != "" && strings.EqualFold(classification.ShortName, classificationDetailName) {
				found = &classification
				break
			}
			// Also check the key (which might be the short name)
			if classificationDetailName != "" && strings.EqualFold(k, classificationDetailName) {
				found = &classification
				break
			}
		}

		if found == nil {
			if classificationDetailID != 0 {
				return fmt.Errorf("classification with ID %d not found", classificationDetailID)
			}
			return fmt.Errorf("classification with name '%s' not found", classificationDetailName)
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
		fmt.Print(format.RenderClassificationDetail(*found, lang))
		return nil
	},
}

func init() {
	classificationDetailCmd.Flags().IntVar(&classificationDetailID, "id", 0, "Classification ID")
	classificationDetailCmd.Flags().StringVar(&classificationDetailName, "name", "", "Classification short name")
	classificationDetailCmd.MarkFlagsOneRequired("id", "name")
	detailCmd.AddCommand(classificationDetailCmd)
	listCmd.AddCommand(classificationsCmd)
}
