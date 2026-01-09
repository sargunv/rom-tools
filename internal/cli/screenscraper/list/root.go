package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List metadata and reference data",
	Long:  "Commands to retrieve lists of metadata and reference data from Screenscraper",
}
