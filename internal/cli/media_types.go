package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var mediaTypesCmd = &cobra.Command{
	Use:   "media-types",
	Short: "Get list of media types",
	Long:  "Retrieves the list of available media types for games and systems",
}

var gameMediaTypesCmd = &cobra.Command{
	Use:   "games",
	Short: "Get list of game media types",
	Long:  "Retrieves the list of available media types for games",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetGameMediaList()
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

var systemMediaTypesCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of system media types",
	Long:  "Retrieves the list of available media types for systems",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSystemMediaList()
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
	mediaTypesCmd.AddCommand(gameMediaTypesCmd)
	mediaTypesCmd.AddCommand(systemMediaTypesCmd)
	rootCmd.AddCommand(mediaTypesCmd)
}
