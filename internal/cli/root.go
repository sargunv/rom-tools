package cli

import (
	"github.com/sargunv/rom-tools/internal/cli/identify"
	"github.com/sargunv/rom-tools/internal/cli/scrape"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rom-tools",
	Short: "ROM management and metadata tools",
	Long:  `A collection of tools for managing ROMs and fetching game metadata.`,
}

func init() {
	rootCmd.AddCommand(identify.Cmd)
	rootCmd.AddCommand(scrape.Cmd)
	rootCmd.AddCommand(screenscraper.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func GetRootCommandForDocs() *cobra.Command {
	return rootCmd
}
