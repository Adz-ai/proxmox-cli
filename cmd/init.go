package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize or reconfigure Proxmox CLI",
	Long: `Initialize the Proxmox CLI configuration.
	
This command helps you set up or reconfigure your connection to a Proxmox VE server.
It will prompt you for the server URL and save it to the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		
		// Check if already configured
		existingURL := viper.GetString("server_url")
		if existingURL != "" && !force {
			fmt.Printf("⚠️  Already configured for server: %s\n", existingURL)
			fmt.Println("Use --force to reconfigure")
			return
		}

		reader := bufio.NewReader(os.Stdin)
		
		fmt.Println("🚀 Proxmox CLI Configuration")
		fmt.Println("============================")
		
		if existingURL != "" {
			fmt.Printf("Current server: %s\n\n", existingURL)
		}
		
		fmt.Print("Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")
		serverURL, _ := reader.ReadString('\n')
		serverURL = strings.TrimSpace(serverURL)

		// Validate input
		if serverURL == "" {
			fmt.Println("❌ Server URL cannot be empty")
			return
		}

		// Ensure URL has protocol
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}

		// Remove /api2/json if user included it
		serverURL = strings.TrimSuffix(serverURL, "/api2/json")
		serverURL = strings.TrimSuffix(serverURL, "/")

		// Save configuration
		viper.Set("server_url", serverURL)
		
		// Clear any existing auth if reconfiguring
		if force && existingURL != "" && existingURL != serverURL {
			viper.Set("auth_ticket", nil)
			fmt.Println("🔄 Cleared existing authentication (server changed)")
		}
		
		err := viper.WriteConfig()
		if err != nil {
			// Try to create config file if it doesn't exist
			configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli")
			os.MkdirAll(configPath, os.ModePerm)
			err = viper.SafeWriteConfig()
			if err != nil {
				fmt.Printf("❌ Failed to save configuration: %s\n", err)
				return
			}
		}

		configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli", "config.json")
		fmt.Printf("\n✅ Configuration saved to %s\n", configPath)
		fmt.Printf("📡 Server URL: %s\n", serverURL)
		fmt.Println("\n📌 Next step: Run 'proxmox-cli auth login -u <username>' to authenticate")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Force reconfiguration even if already configured")
}