package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Authenticate with a Proxmox API token",
		Long: `Store a Proxmox API token for authentication.

Unlike session tickets, API tokens do not expire, which makes them the
recommended way to use this CLI in scripts. Create one in the Proxmox web
interface under Datacenter > Permissions > API Tokens, then run:

  proxmox-cli auth token -t 'user@realm!tokenname'

The token secret is prompted for interactively (or read from stdin when
piped) so it does not end up in your shell history.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			tokenID, err := cmd.Flags().GetString("token-id")
			if err != nil {
				return fmt.Errorf("read token-id flag: %w", err)
			}
			tokenID = strings.TrimSpace(tokenID)
			if !strings.Contains(tokenID, "!") {
				return fmt.Errorf("token ID must have the form 'user@realm!tokenname'")
			}

			secret, err := promptSecret(cmd, "Enter API token secret: ")
			if err != nil {
				return err
			}
			if secret == "" {
				return fmt.Errorf("token secret cannot be empty")
			}

			serverURL := strings.TrimSpace(utility.ContextString("server_url"))
			if serverURL == "" {
				return fmt.Errorf("server URL is not configured; run 'proxmox-cli init'")
			}
			serverURL, err = utility.NormalizeServerURL(serverURL)
			if err != nil {
				return fmt.Errorf("invalid server URL: %w", err)
			}

			fmt.Fprintf(out, "Verifying API token against %s...\n", serverURL)

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			httpClient, err := utility.NewHTTPClient(utility.ContextBool("insecure"), utility.ContextString("ca_cert"))
			if err != nil {
				return fmt.Errorf("configure HTTP client: %w", err)
			}
			client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", serverURL),
				proxmox.WithHTTPClient(httpClient),
				proxmox.WithAPIToken(tokenID, secret),
			)
			version, err := client.Version(ctx)
			if err != nil {
				return fmt.Errorf("verify API token: %w", err)
			}

			utility.ClearAuthTicket()
			utility.ClearAPIToken()
			utility.SetContextValue("api_token.token_id", tokenID)
			utility.SetContextValue("api_token.secret", secret)
			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("save API token: %w", err)
			}

			fmt.Fprintln(out, "API token saved")
			fmt.Fprintf(out, "Connected to Proxmox VE %s\n", version.Version)
			return nil
		},
	}

	cmd.Flags().StringP("token-id", "t", "", "API token ID in the form 'user@realm!tokenname' (required)")
	if err := cmd.MarkFlagRequired("token-id"); err != nil {
		panic(err)
	}
	return cmd
}
