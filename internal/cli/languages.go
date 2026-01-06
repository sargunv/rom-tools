package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var languagesCmd = &cobra.Command{
	Use:   "languages",
	Short: "Get list of languages",
	Long:  "Retrieves the list of all languages",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetLanguagesList()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(languagesCmd)
}
