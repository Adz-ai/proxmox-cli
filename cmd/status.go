package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"proxmox-cli/cmd/utility"
	"time"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration and connection status",
	Long:  `Display the current Proxmox CLI configuration, authentication status, and test the connection.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Proxmox CLI Status")
		fmt.Println("====================")
		
		// Show configuration path
		configPath := filepath.Join(os.Getenv("HOME"), ".proxmox-cli", "config.json")
		fmt.Printf("\n📁 Config file: %s\n", configPath)
		
		// Check server URL
		serverURL := viper.GetString("server_url")
		if serverURL == "" {
			fmt.Println("\n❌ Status: Not configured")
			fmt.Println("💡 Run 'proxmox-cli init' to configure")
			return
		}
		
		fmt.Printf("\n🖥️  Server URL: %s\n", serverURL)
		
		// Check authentication
		authTicket := viper.Sub("auth_ticket")
		if authTicket == nil || authTicket.GetString("ticket") == "" {
			fmt.Println("🔐 Authentication: Not logged in")
			fmt.Println("💡 Run 'proxmox-cli auth login -u <username>' to authenticate")
			return
		}
		
		fmt.Println("🔐 Authentication: Logged in ✓")
		
		// Test connection
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Println("\n🔄 Testing connection...")
			
			client := utility.GetClient()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			version, err := client.Version(ctx)
			if err != nil {
				fmt.Printf("❌ Connection failed: %s\n", err)
				fmt.Println("💡 Try running 'proxmox-cli auth login -u <username>' to re-authenticate")
			} else {
				fmt.Println("✅ Connection successful!")
				fmt.Printf("📊 Proxmox VE Version: %s\n", version.Version)
				fmt.Printf("📦 Release: %s\n", version.Release)
				
				// Show cluster info
				nodes, err := client.Nodes(ctx)
				if err == nil {
					fmt.Printf("\n🌐 Cluster nodes: %d\n", len(nodes))
					for _, node := range nodes {
						status := "🔴 offline"
						if node.Status == "online" {
							status = "🟢 online"
						}
						fmt.Printf("   - %s %s\n", node.Node, status)
					}
				}
			}
		} else {
			fmt.Println("\n💡 Use --verbose to test the connection")
		}
		
		fmt.Println("\n📚 Available commands:")
		fmt.Println("   - proxmox-cli nodes get     # List cluster nodes")
		fmt.Println("   - proxmox-cli vm get        # List virtual machines")
		fmt.Println("   - proxmox-cli lxc get       # List containers")
		fmt.Println("   - proxmox-cli --help        # Show all commands")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolP("verbose", "v", false, "Test connection and show detailed information")
}