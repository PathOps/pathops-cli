package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pathops",
		Short: "PathOps CLI",
		Long:  "PathOps CLI talks to the PathOps Control Plane.",
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newWhoamiCmd())

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
