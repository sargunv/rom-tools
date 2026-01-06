package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var romTypesCmd = &cobra.Command{
	Use:   "rom-types",
	Short: "Get list of ROM types",
	Long:  "Retrieves the list of all ROM types",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetROMTypesList()
		if err != nil {
			return err
		}

		formatted, err := json.MarshalIndent(resp.Response.ROMTypes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

func init() {
	listCmd.AddCommand(romTypesCmd)
}
