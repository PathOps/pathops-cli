package cli

import (
	"fmt"

	"github.com/pathops/pathops-cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pathops %s\ncommit=%s\ndate=%s\n",
				version.Version,
				version.Commit,
				version.Date,
			)
		},
	}
}
