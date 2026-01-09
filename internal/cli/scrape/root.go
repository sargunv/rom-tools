package scrape

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape metadata for ROM collections",
	Long:  `Batch scrape metadata and media for ROM files in a directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement scraping logic
		return cmd.Help()
	},
}
