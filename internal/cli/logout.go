package cli

import (
	"fmt"

	"github.com/pathops/pathops-cli/internal/auth"
	"github.com/pathops/pathops-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out from PathOps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			profile, err := config.Active(cfg)
			if err != nil {
				return err
			}

			logoutURL, err := auth.BuildLogoutURL(profile.Issuer)
			if err != nil {
				return err
			}

			if err := auth.DeleteTokens(); err != nil {
				return err
			}

			fmt.Println("Local PathOps session removed.")

			if err := auth.OpenBrowser(logoutURL); err != nil {
				fmt.Printf("Could not open browser automatically.\nOpen this URL manually to complete logout:\n%s\n", logoutURL)
				return nil
			}

			fmt.Println("Opened browser to complete Keycloak logout.")
			fmt.Printf("If needed, open this URL manually:\n%s\n", logoutURL)
			return nil
		},
	}
}

func chooseRefreshToken(newValue, oldValue string) string {
	if newValue != "" {
		return newValue
	}
	return oldValue
}
