package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize or reconfigure Proxmox CLI",
	Long: `Initialize the Proxmox CLI configuration.
	
This command helps you set up or reconfigure your connection to a Proxmox VE server.
It will prompt you for the server URL and save it to the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()
		force, _ := cmd.Flags().GetBool("force")

		// Check if already configured
		existingURL := viper.GetString("server_url")
		if existingURL != "" && !force {
			fmt.Fprintf(out, "⚠️  Already configured for server: %s\n", existingURL)
			fmt.Fprintln(out, "Use --force to reconfigure")
			return
		}

		reader := bufio.NewReader(in)

		fmt.Fprintln(out, "Welcome to Proxmox CLI")
		fmt.Fprintln(out, "🚀 Proxmox CLI Configuration")
		fmt.Fprintln(out, "============================")

		if existingURL != "" {
			fmt.Fprintf(out, "Current server: %s\n\n", existingURL)
		}

		fmt.Fprint(out, "Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")
		
		// For testing, we need to handle the case where stdin is not a terminal
		serverURL, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(out, "❌ Error reading input: %s\n", err)
			return
		}
		serverURL = strings.TrimSpace(serverURL)

		// Validate input
		if serverURL == "" {
			fmt.Fprintln(out, "❌ Server URL cannot be empty")
			return
		}

		// Ensure URL has protocol
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}

		// Remove /api2/json if the user included it
		serverURL = strings.TrimSuffix(serverURL, "/api2/json")
		serverURL = strings.TrimSuffix(serverURL, "/")

		// Save configuration
		viper.Set("server_url", serverURL)

		// Clear any existing auth if reconfiguring
		if force && existingURL != "" && existingURL != serverURL {
			viper.Set("auth_ticket", nil)
			fmt.Fprintln(out, "🔄 Cleared existing authentication (server changed)")
		}

		err = viper.WriteConfig()
		if err != nil {
			// Try to create a config file if it doesn't exist
			configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli")
			err := os.MkdirAll(configPath, os.ModePerm)
			if err != nil {
				fmt.Fprintf(out, "❌ Failed to create configuration directory: %s\n", err)
				return
			}
			err = viper.SafeWriteConfig()
			if err != nil {
				fmt.Fprintf(out, "❌ Failed to save configuration: %s\n", err)
				return
			}
		}

		configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli", "config.json")
		fmt.Fprintf(out, "\n✅ Configuration saved to %s\n", configPath)
		fmt.Fprintf(out, "📡 Server URL: %s\n", serverURL)
		fmt.Fprintln(out, "\n📌 Next step: Run 'proxmox-cli auth login -u <username>' to authenticate")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Force reconfiguration even if already configured")
}