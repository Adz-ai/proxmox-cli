package auth

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"net/http"
	"os"
	"strings"
)

// viewCmd represents the view subcommand
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Proxmox and retrieve an auth cookie",
	Long:  `Authenticate with Proxmox by providing a username and password, and retrieve an authentication cookie to use for subsequent API requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		password := getPassword()

		// Call function to handle authentication
		if username == "" || password == "" {
			fmt.Println("Both username and password are required.")
			return
		}

		authenticateWithProxmox(username, password)
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username for Proxmox (required)")
	err := loginCmd.MarkFlagRequired("username")
	if err != nil {
		return
	}
}

func getPassword() string {
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nFailed to read password")
		return ""
	}
	fmt.Println() // It's common to output a newline after reading a password
	return strings.TrimSpace(string(bytePassword))
}

func authenticateWithProxmox(username, password string) {
	// Ensure the server URL is configured
	serverURL := viper.GetString("server_url")
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		// Prompt for server URL if not configured
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("🔧 Proxmox server URL not configured.")
		fmt.Print("Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")
		serverURL, _ = reader.ReadString('\n')
		serverURL = strings.TrimSpace(serverURL)

		// Ensure URL has protocol
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}

		// Remove /api2/json if user included it
		serverURL = strings.TrimSuffix(serverURL, "/api2/json")
		serverURL = strings.TrimSuffix(serverURL, "/")

		// Save the URL
		viper.Set("server_url", serverURL)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Printf("❌ Failed to save configuration: %s\n", err)
			return
		}
		fmt.Println("✅ Server URL saved to configuration")
		fmt.Println()
	}

	// Show which server we're connecting to
	fmt.Printf("🔐 Authenticating with Proxmox server at %s...\n", serverURL)

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
		fmt.Printf("❌ Failed to authenticate: %s\n", err)
		fmt.Println("\n💡 Common issues:")
		fmt.Println("  - Check username format (e.g., root@pam, user@pve)")
		fmt.Println("  - Verify the server URL is correct")
		fmt.Println("  - Ensure your password is correct")
		fmt.Println("  - Check if the Proxmox server is accessible")
		return
	}

	// Get and store the ticket and CSRF token
	ticket, err := client.Ticket(context.Background(), &credentials)
	if err != nil {
		fmt.Printf("❌ Failed to get authentication ticket: %s\n", err)
		return
	}

	viper.Set("auth_ticket", ticket)
	err = viper.WriteConfig()
	if err != nil {
		fmt.Printf("❌ Failed to save authentication: %s\n", err)
		return
	}

	fmt.Printf("✅ Authentication successful!\n")
	fmt.Printf("📊 Connected to Proxmox VE %s\n", version.Version)
	fmt.Println("\n🎯 You can now use commands like:")
	fmt.Println("  - proxmox-cli nodes get")
	fmt.Println("  - proxmox-cli vm get")
	fmt.Println("  - proxmox-cli lxc get")
}
