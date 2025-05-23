package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"proxmox-cli/cmd/utility"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration and connection status",
	Long:  `Display the current Proxmox CLI configuration, authentication status, and test the connection.`,
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		fmt.Fprintln(out, "🔍 Proxmox CLI Status")
		fmt.Fprintln(out, "====================")

		// Show the configuration path
		configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli", "config.json")
		fmt.Fprintf(out, "\n📁 Config file: %s\n", configPath)

		// Check server URL
		serverURL := viper.GetString("server_url")
		if serverURL == "" {
			fmt.Fprintln(out, "\n❌ Status: Not configured")
			fmt.Fprintln(out, "💡 Run 'proxmox-cli init' to configure")
			return
		}

		fmt.Fprintf(out, "\n🖥️  Server URL: %s\n", serverURL)

		// Check authentication
		authTicket := viper.Sub("auth_ticket")
		if authTicket == nil || authTicket.GetString("ticket") == "" {
			fmt.Fprintln(out, "🔐 Authentication: Not logged in")
			fmt.Fprintln(out, "💡 Run 'proxmox-cli auth login -u <username>' to authenticate")
			return
		}

		fmt.Fprintln(out, "🔐 Authentication: Logged in ✓")

		// Test connection
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Fprintln(out, "\n🔄 Testing connection...")

			client := utility.GetClient()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			version, err := client.Version(ctx)
			if err != nil {
				fmt.Fprintf(out, "❌ Connection failed: %s\n", err)
				fmt.Fprintln(out, "💡 Try running 'proxmox-cli auth login -u <username>' to re-authenticate")
			} else {
				fmt.Fprintln(out, "✅ Connection successful!")
				fmt.Fprintf(out, "📊 Proxmox VE Version: %s\n", version.Version)
				fmt.Fprintf(out, "📦 Release: %s\n", version.Release)

				// Show cluster info
				nodes, err := client.Nodes(ctx)
				if err == nil {
					fmt.Fprintf(out, "\n🌐 Cluster nodes: %d\n", len(nodes))
					for _, node := range nodes {
						status := "🔴 offline"
						if node.Status == "online" {
							status = "🟢 online"
						}
						fmt.Fprintf(out, "   - %s %s\n", node.Node, status)
					}
				}
			}
		} else {
			fmt.Fprintln(out, "\n💡 Use --verbose to test the connection")
		}

		fmt.Fprintln(out, "\n📚 Available commands:")
		fmt.Fprintln(out, "   - proxmox-cli nodes get     # List cluster nodes")
		fmt.Fprintln(out, "   - proxmox-cli vm get        # List virtual machines")
		fmt.Fprintln(out, "   - proxmox-cli lxc get       # List containers")
		fmt.Fprintln(out, "   - proxmox-cli --help        # Show all commands")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolP("verbose", "v", false, "Test connection and show detailed information")
}