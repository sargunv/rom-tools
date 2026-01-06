package cli

import (
	"fmt"
	"os"

	screenscraper "sargunv/screenscraper-go/client"

	"github.com/spf13/cobra"
)

var (
	devID       string
	devPassword string
	softName    string
	ssID        string
	ssPassword  string
	client      *screenscraper.Client
)

var rootCmd = &cobra.Command{
	Use:   "screenscraper",
	Short: "Screenscraper API CLI client",
	Long: `A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:
  SCREENSCRAPER_DEV_USER     - Developer username
  SCREENSCRAPER_DEV_PASSWORD - Developer password
  SCREENSCRAPER_ID           - User ID (optional)
  SCREENSCRAPER_PASSWORD     - User password (optional)`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize client with credentials from environment or flags
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

		if softName == "" {
			softName = "screenscraper-go"
		}

		client = screenscraper.NewClient(devID, devPassword, softName, ssID, ssPassword)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&devID, "dev-id", "", "Developer ID (or set SCREENSCRAPER_DEV_USER)")
	rootCmd.PersistentFlags().StringVar(&devPassword, "dev-password", "", "Developer password (or set SCREENSCRAPER_DEV_PASSWORD)")
	rootCmd.PersistentFlags().StringVar(&ssID, "user-id", "", "User ID (or set SCREENSCRAPER_ID)")
	rootCmd.PersistentFlags().StringVar(&ssPassword, "user-password", "", "User password (or set SCREENSCRAPER_PASSWORD)")
	rootCmd.PersistentFlags().StringVar(&softName, "soft-name", "", "Software name identifier")
}

func Execute() error {
	return rootCmd.Execute()
}
