package screenscraper

import (
	"fmt"
	"os"

	"github.com/sargunv/rom-tools/clients/screenscraper"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/detail"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/list"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/status"

	"github.com/spf13/cobra"
)

var (
	devID       string
	devPassword string
	ssID        string
	ssPassword  string
)

var Cmd = &cobra.Command{
	Use:   "screenscraper",
	Short: "Screenscraper API client",
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

		shared.Client = screenscraper.NewClient(devID, devPassword, "screenscraper-go", ssID, ssPassword)
	},
}

func init() {
	Cmd.PersistentFlags().BoolVar(&shared.JsonOutput, "json", false, "Output results as JSON")
	Cmd.PersistentFlags().StringVar(&shared.Locale, "locale", "", "Override locale for output (e.g., en, fr, de)")

	// Add sub-package commands
	Cmd.AddCommand(detail.Cmd)
	Cmd.AddCommand(list.Cmd)
	Cmd.AddCommand(status.Cmd)
}
