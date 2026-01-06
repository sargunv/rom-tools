package cli

import (
	"encoding/json"
	"fmt"

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

		formatted, err := json.MarshalIndent(resp.Response.Classifications, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	listCmd.AddCommand(classificationsCmd)
}
