package auth

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Proxmox and retrieve an auth cookie",
		Long:  `Authenticate with Proxmox by providing a username and password, and retrieve an authentication cookie to use for subsequent API requests.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := cmd.Flags().GetString("username")
			if err != nil {
				return fmt.Errorf("read username: %w", err)
			}
			password, err := promptSecret(cmd, "Enter Password: ")
			if err != nil {
				return err
			}
			if username == "" || password == "" {
				return fmt.Errorf("both username and password are required")
			}

			return authenticateWithProxmox(cmd, username, password)
		},
	}

	cmd.Flags().StringP("username", "u", "", "Username for Proxmox (required)")
	if err := cmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	return cmd
}

// promptSecret reads a secret without echoing when stdin is a terminal, and
// falls back to a plain read for piped input.
func promptSecret(cmd *cobra.Command, prompt string) (string, error) {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	if file, ok := in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(out, prompt)
		byteSecret, err := term.ReadPassword(int(file.Fd()))
		if err != nil {
			return "", fmt.Errorf("read secret: %w", err)
		}
		fmt.Fprintln(out)
		return strings.TrimSpace(string(byteSecret)), nil
	}

	reader := bufio.NewReader(in)
	secret, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read secret: %w", err)
	}
	return strings.TrimSpace(secret), nil
}

func authenticateWithProxmox(cmd *cobra.Command, username, password string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	serverURL := strings.TrimSpace(viper.GetString("server_url"))
	configuredURL := serverURL
	if serverURL == "" {
		reader := bufio.NewReader(in)
		fmt.Fprintln(out, "Proxmox server URL not configured.")
		fmt.Fprint(out, "Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")

		var err error
		serverURL, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read server URL: %w", err)
		}
	}

	serverURL, err := utility.NormalizeServerURL(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	if configuredURL != serverURL {
		viper.Set("server_url", serverURL)
		if err := utility.WriteConfig(); err != nil {
			return fmt.Errorf("save server URL: %w", err)
		}
		fmt.Fprintln(out, "Server URL saved to configuration")
	}

	fmt.Fprintf(out, "Authenticating with Proxmox server at %s...\n", serverURL)

	// Start the timeout only once all interactive prompting is done, so slow
	// typing does not count against the API deadline.
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	httpClient, err := utility.NewHTTPClient(viper.GetBool("insecure"), viper.GetString("ca_cert"))
	if err != nil {
		return fmt.Errorf("configure HTTP client: %w", err)
	}

	credentials := proxmox.Credentials{
		Username: username,
		Password: password,
	}
	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", serverURL),
		proxmox.WithHTTPClient(httpClient),
	)

	ticket, err := client.Ticket(ctx, &credentials)
	if err != nil {
		return fmt.Errorf("authenticate with Proxmox: %w", err)
	}

	version, err := client.Version(ctx)
	if err != nil {
		return fmt.Errorf("get Proxmox version: %w", err)
	}

	// Store only the credentials the CLI needs; the session also carries
	// username, capabilities, and cluster name that don't belong in the config.
	// Any stored API token is cleared so the fresh login takes effect.
	utility.ClearAuthTicket()
	utility.ClearAPIToken()
	viper.Set("auth_ticket.ticket", ticket.Ticket)
	viper.Set("auth_ticket.CSRFPreventionToken", ticket.CSRFPreventionToken)
	if err := utility.WriteConfig(); err != nil {
		return fmt.Errorf("save authentication: %w", err)
	}

	fmt.Fprintln(out, "Authentication successful")
	fmt.Fprintf(out, "Connected to Proxmox VE %s\n", version.Version)
	fmt.Fprintln(out, "\nYou can now use commands like:")
	fmt.Fprintln(out, "  - proxmox-cli nodes get")
	fmt.Fprintln(out, "  - proxmox-cli vm get")
	fmt.Fprintln(out, "  - proxmox-cli lxc get")
	return nil
}
