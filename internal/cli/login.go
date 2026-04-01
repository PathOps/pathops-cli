package cli

import (
	"fmt"

	"github.com/pathops/pathops-cli/internal/api"
	"github.com/pathops/pathops-cli/internal/auth"
	"github.com/pathops/pathops-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with PathOps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			profile, err := config.Active(cfg)
			if err != nil {
				return err
			}

			fmt.Printf("Starting login against %s\n", profile.Issuer)

			result, err := auth.LoginWithBrowserPKCE(profile.Issuer, profile.ClientID)
			if err != nil {
				return err
			}

			if err := auth.SaveTokens(auth.Tokens{
				AccessToken:  result.AccessToken,
				RefreshToken: result.RefreshToken,
				IDToken:      result.IDToken,
				TokenType:    "Bearer",
			}); err != nil {
				return err
			}

			client := api.New(profile.ControlPlaneBaseURL)
			loginResp, err := client.PublicLogin(result.AccessToken)
			if err != nil {
				return err
			}

			fmt.Printf("Logged in.\nTenant: %s (%s)\nRole: %s\n",
				loginResp.Data.TenantName,
				loginResp.Data.TenantSlug,
				loginResp.Data.MembershipRole,
			)

			if loginResp.Data.RequiresTokenRefresh {
				fmt.Println("Token refresh required.")
			}
			if loginResp.Data.RequiresToolRelogin {
				fmt.Println("Tool relogin required.")
			}

			return nil
		},
	}
}