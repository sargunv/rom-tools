package cli

import (
	"github.com/spf13/cobra"
)

var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "Get detailed information about a specific item",
	Long:  "Commands to retrieve detailed information about systems, genres, families, classifications, and games.",
}

func init() {
	rootCmd.AddCommand(detailCmd)
}
