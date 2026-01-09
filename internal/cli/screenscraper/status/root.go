package status

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Get status information",
	Long:  "Commands to retrieve status information about the API infrastructure and user account.",
}
