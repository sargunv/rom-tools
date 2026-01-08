package cli

import (
	"fmt"
	"os"

	screenscraper "github.com/sargunv/screenscraper-go/client"

	"github.com/spf13/cobra"
)

var (
	devID       string
	devPassword string
	ssID        string
	ssPassword  string
	jsonOutput  bool
	locale      string
	client      *screenscraper.Client
)

var rootCmd = &cobra.Command{
	Use:   "screenscraper",
	Short: "Screenscraper CLI client",
	Long: `A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:

- SCREENSCRAPER_DEV_USER     - Developer username
- SCREENSCRAPER_DEV_PASSWORD - Developer password
- SCREENSCRAPER_ID           - User ID (optional)
- SCREENSCRAPER_PASSWORD     - User password (optional)`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize client with credentials from environment variables
		if devID == "" {
			devID = os.Getenv("SCREENSCRAPER_DEV_USER")
		}
		if devPassword == "" {
			devPassword = os.Getenv("SCREENSCRAPER_DEV_PASSWORD")
		}
		if ssID == "" {
			ssID = os.Getenv("SCREENSCRAPER_ID")
		}
		if ssPassword == "" {
			ssPassword = os.Getenv("SCREENSCRAPER_PASSWORD")
		}

		if devID == "" || devPassword == "" {
			fmt.Fprintln(os.Stderr, "Error: Developer credentials required")
			fmt.Fprintln(os.Stderr, "Set SCREENSCRAPER_DEV_USER and SCREENSCRAPER_DEV_PASSWORD environment variables")
			os.Exit(1)
		}

		client = screenscraper.NewClient(devID, devPassword, "screenscraper-go", ssID, ssPassword)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	rootCmd.PersistentFlags().StringVar(&locale, "locale", "", "Override locale for output (e.g., en, fr, de)")
}

func Execute() error {
	return rootCmd.Execute()
}

func GetRootCommandForDocs() *cobra.Command {
	return rootCmd
}
