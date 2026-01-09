package detail

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "detail",
	Short: "Get detailed information about a specific item",
	Long:  "Commands to retrieve detailed information about systems and games.",
}
