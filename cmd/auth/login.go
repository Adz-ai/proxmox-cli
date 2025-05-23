package auth

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// viewCmd represents the view subcommand
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Proxmox and retrieve an auth cookie",
	Long:  `Authenticate with Proxmox by providing a username and password, and retrieve an authentication cookie to use for subsequent API requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		username, _ := cmd.Flags().GetString("username")
		password := getPassword(cmd)

		// Call function to handle authentication
		if username == "" || password == "" {
			fmt.Fprintln(out, "Both username and password are required.")
			return
		}

		authenticateWithProxmox(cmd, username, password)
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username for Proxmox (required)")
	err := loginCmd.MarkFlagRequired("username")
	if err != nil {
		return
	}
}

func getPassword(cmd *cobra.Command) string {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	
	// Check if stdin is a terminal
	if file, ok := in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(out, "Enter Password: ")
		bytePassword, err := term.ReadPassword(int(file.Fd()))
		if err != nil {
			fmt.Fprintln(out, "\nFailed to read password")
			return ""
		}
		fmt.Fprintln(out) // It's common to output a newline after reading a password
		return strings.TrimSpace(string(bytePassword))
	} else {
		// For non-terminal input (like in tests), read from stdin
		reader := bufio.NewReader(in)
		password, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(out, "Failed to read password: %v\n", err)
			return ""
		}
		return strings.TrimSpace(password)
	}
}

func authenticateWithProxmox(cmd *cobra.Command, username, password string) {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	
	// Ensure the server URL is configured
	serverURL := viper.GetString("server_url")
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		// Prompt for server URL if is not configured
		reader := bufio.NewReader(in)
		fmt.Fprintln(out, "🔧 Proxmox server URL not configured.")
		fmt.Fprint(out, "Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")
		
		var err error
		serverURL, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(out, "❌ Error reading server URL: %s\n", err)
			return
		}
		serverURL = strings.TrimSpace(serverURL)

		// Ensure URL has protocol
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}

		// Remove /api2/json if the user included it
		serverURL = strings.TrimSuffix(serverURL, "/api2/json")
		serverURL = strings.TrimSuffix(serverURL, "/")

		// Save the URL
		viper.Set("server_url", serverURL)
		err = viper.WriteConfig()
		if err != nil {
			fmt.Fprintf(out, "❌ Failed to save configuration: %s\n", err)
			return
		}
		fmt.Fprintln(out, "✅ Server URL saved to configuration")
		fmt.Fprintln(out)
	}

	// Show which server we're connecting to
	fmt.Fprintf(out, "🔐 Authenticating with Proxmox server at %s...\n", serverURL)

	// Configure HTTP client with TLS settings
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Allow self-signed certificates
			},
		},
	}

	// Authenticate with Proxmox
	credentials := proxmox.Credentials{
		Username: username,
		Password: password,
	}
	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", serverURL),
		proxmox.WithCredentials(&credentials),
		proxmox.WithHTTPClient(httpClient),
	)

	// Get the version to test connection
	version, err := client.Version(context.Background())
	if err != nil {
		fmt.Fprintf(out, "❌ Failed to authenticate: %s\n", err)
		fmt.Fprintln(out, "\n💡 Common issues:")
		fmt.Fprintln(out, "  - Check username format (e.g., root@pam, user@pve)")
		fmt.Fprintln(out, "  - Verify the server URL is correct")
		fmt.Fprintln(out, "  - Ensure your password is correct")
		fmt.Fprintln(out, "  - Check if the Proxmox server is accessible")
		return
	}

	// Get and store the ticket and CSRF token
	ticket, err := client.Ticket(context.Background(), &credentials)
	if err != nil {
		fmt.Fprintf(out, "❌ Failed to get authentication ticket: %s\n", err)
		return
	}

	viper.Set("auth_ticket", ticket)
	err = viper.WriteConfig()
	if err != nil {
		fmt.Fprintf(out, "❌ Failed to save authentication: %s\n", err)
		return
	}

	fmt.Fprintln(out, "✅ Authentication successful!")
	fmt.Fprintf(out, "📊 Connected to Proxmox VE %s\n", version.Version)
	fmt.Fprintln(out, "\n🎯 You can now use commands like:")
	fmt.Fprintln(out, "  - proxmox-cli nodes get")
	fmt.Fprintln(out, "  - proxmox-cli vm get")
	fmt.Fprintln(out, "  - proxmox-cli lxc get")
}