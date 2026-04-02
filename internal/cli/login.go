package cli

import (
	"context"
	"fmt"

	"github.com/pathops/pathops-cli/internal/api"
	"github.com/pathops/pathops-cli/internal/auth"
	"github.com/pathops/pathops-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var controlPlaneURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with PathOps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			var profile config.Profile

			if controlPlaneURL != "" {
				client := api.New(controlPlaneURL)

				authCfg, err := client.GetPublicAuthConfig()
				if err != nil {
					return err
				}

				profile = config.Profile{
					ControlPlaneBaseURL: authCfg.ControlPlaneBaseURL,
					Issuer:              authCfg.Issuer,
					ClientID:            authCfg.ClientID,
				}

				if err := config.SaveActiveProfile(cfg, profile); err != nil {
					return err
				}
			} else {
				var err error
				profile, err = config.Active(cfg)
				if err != nil {
					return err
				}
			}

			if profile.ControlPlaneBaseURL == "" || profile.Issuer == "" || profile.ClientID == "" {
				return fmt.Errorf("missing profile configuration; run: pathops login --control-plane <url>")
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
				TokenType:    result.TokenType,
				Expiry:       result.Expiry,
				Issuer:       profile.Issuer,
				ClientID:     profile.ClientID,
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
				fmt.Println("Refreshing token because tenant context changed...")

				if result.RefreshToken == "" {
					return fmt.Errorf("control plane requested token refresh, but no refresh token was returned")
				}

				refreshed, err := auth.RefreshToken(
					context.Background(),
					profile.Issuer,
					profile.ClientID,
					result.RefreshToken,
				)
				if err != nil {
					return fmt.Errorf("refresh after login failed: %w", err)
				}

				if err := auth.SaveTokens(auth.Tokens{
					AccessToken:  refreshed.AccessToken,
					RefreshToken: chooseRefreshToken(refreshed.RefreshToken, result.RefreshToken),
					IDToken:      refreshed.IDToken,
					TokenType:    refreshed.TokenType,
					Expiry:       refreshed.Expiry,
					Issuer:       profile.Issuer,
					ClientID:     profile.ClientID,
				}); err != nil {
					return err
				}

				fmt.Println("Token refreshed successfully.")
			}

			if loginResp.Data.RequiresToolRelogin {
				fmt.Println("Some tools may require a new login to pick up tenant changes.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&controlPlaneURL, "control-plane", "", "PathOps Control Plane base URL")
	return cmd
}
