package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Get user information and quotas",
	Long:  "Retrieves user information including quotas, rate limits, and contribution stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetUserInfo()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.SSUser, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderUser(resp.Response.SSUser, lang))
		return nil
	},
}

func init() {
	statusCmd.AddCommand(userCmd)
}
