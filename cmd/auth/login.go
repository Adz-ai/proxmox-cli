package auth

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
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
	Cmd.AddCommand(loginCmd)
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
		fmt.Println("Server URL is not configured. Please run the setup.")
		return
	}

	// Authenticate with Proxmox
	credentials := proxmox.Credentials{
		Username: username,
		Password: password,
	}
	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", serverURL),
		proxmox.WithCredentials(&credentials),
	)

	// Get the version
	_, err := client.Version(context.Background())
	if err != nil {
		fmt.Printf("Failed to authenticate: %s\n", err)
		return
	}

	// Optionally store the ticket and CSRF token in Viper for future use
	ticket, err := client.Ticket(context.Background(), &credentials)
	viper.Set("auth_ticket", ticket)
	err = viper.WriteConfig()
	if err != nil {
		fmt.Printf("Failed to WriteConfig: %s\n", err)
		return
	}

	fmt.Println("Authentication successful, ticket stored.")
}
