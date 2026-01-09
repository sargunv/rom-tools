package screenscraper

import (
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/clients/screenscraper"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"

	"github.com/spf13/cobra"
)

var (
	rateID     string
	rateRating int
)

var rateCmd = &cobra.Command{
	Use:   "rate",
	Short: "Submit a rating for a game",
	Long: `Submit a rating (1-20) for a game on ScreenScraper.

This command requires user credentials (SCREENSCRAPER_ID and SCREENSCRAPER_PASSWORD).
Your rating will be associated with your ScreenScraper account.`,
	Example: `  # Rate a game 18 out of 20
  rom-tools screenscraper rate --id=3 --rating=18`,
	RunE: func(cmd *cobra.Command, args []string) error {
		params := screenscraper.SubmitRatingParams{
			GameID: rateID,
			Rating: rateRating,
		}

		resp, err := shared.Client.SubmitRating(params)
		if err != nil {
			return err
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		fmt.Println(resp.Message)
		return nil
	},
}

func init() {
	rateCmd.Flags().StringVar(&rateID, "id", "", "Game ID to rate (required)")
	rateCmd.Flags().IntVar(&rateRating, "rating", 0, "Rating from 1 to 20 (required)")

	rateCmd.MarkFlagRequired("id")
	rateCmd.MarkFlagRequired("rating")

	Cmd.AddCommand(rateCmd)
}
