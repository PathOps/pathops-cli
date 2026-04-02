package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pathops/pathops-cli/internal/api"
	"github.com/pathops/pathops-cli/internal/auth"
	"github.com/pathops/pathops-cli/internal/config"
	"github.com/spf13/cobra"
)

func newWhoamiCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current CLI identity context",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			profile, err := config.Active(cfg)
			if err != nil {
				return err
			}

			out := api.WhoAmI{
				Profile:   cfg.ActiveProfile,
				BaseURL:   profile.ControlPlaneBaseURL,
				HasTokens: auth.HasTokens(),
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("profile: %s\nbaseUrl: %s\nhasTokens: %v\n",
				out.Profile,
				out.BaseURL,
				out.HasTokens,
			)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as json")
	return cmd
}
