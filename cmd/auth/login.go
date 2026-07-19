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
			password, err := getPassword(cmd)
			if err != nil {
				return err
			}
			if username == "" || password == "" {
				return fmt.Errorf("both username and password are required")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			return authenticateWithProxmox(ctx, cmd, username, password)
		},
	}

	cmd.Flags().StringP("username", "u", "", "Username for Proxmox (required)")
	if err := cmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	return cmd
}

func getPassword(cmd *cobra.Command) (string, error) {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	if file, ok := in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(out, "Enter Password: ")
		bytePassword, err := term.ReadPassword(int(file.Fd()))
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		fmt.Fprintln(out)
		return strings.TrimSpace(string(bytePassword)), nil
	}

	reader := bufio.NewReader(in)
	password, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimSpace(password), nil
}

func authenticateWithProxmox(ctx context.Context, cmd *cobra.Command, username, password string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	serverURL := strings.TrimSpace(viper.GetString("server_url"))
	configuredURL := serverURL
	if serverURL == "" {
		reader := bufio.NewReader(in)
		fmt.Fprintln(out, "🔧 Proxmox server URL not configured.")
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
		fmt.Fprintln(out, "✅ Server URL saved to configuration")
	}

	fmt.Fprintf(out, "🔐 Authenticating with Proxmox server at %s...\n", serverURL)

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

	viper.Set("auth_ticket", ticket)
	if err := utility.WriteConfig(); err != nil {
		return fmt.Errorf("save authentication: %w", err)
	}

	fmt.Fprintln(out, "✅ Authentication successful!")
	fmt.Fprintf(out, "📊 Connected to Proxmox VE %s\n", version.Version)
	fmt.Fprintln(out, "\n🎯 You can now use commands like:")
	fmt.Fprintln(out, "  - proxmox-cli nodes get")
	fmt.Fprintln(out, "  - proxmox-cli vm get")
	fmt.Fprintln(out, "  - proxmox-cli lxc get")
	return nil
}
