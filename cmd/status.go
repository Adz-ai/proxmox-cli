package cmd

import (
	"context"
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current configuration and connection status",
		Long:  `Display the current Proxmox CLI configuration, authentication status, and test the connection.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, "Proxmox CLI Status")
			fmt.Fprintln(out, "==================")

			configPath, err := utility.ConfigFile()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "\nConfig file: %s\n", configPath)

			// Check server URL
			serverURL := viper.GetString("server_url")
			if serverURL == "" {
				fmt.Fprintln(out, "\nStatus: Not configured")
				fmt.Fprintln(out, "Run 'proxmox-cli init' to configure")
				return nil
			}

			fmt.Fprintf(out, "\nServer URL: %s\n", serverURL)

			// Check authentication
			switch {
			case utility.HasAPIToken():
				fmt.Fprintln(out, "Authentication: Logged in (API token)")
			case utility.HasSessionTicket():
				fmt.Fprintln(out, "Authentication: Logged in (session ticket)")
			default:
				fmt.Fprintln(out, "Authentication: Not logged in")
				fmt.Fprintln(out, "Run 'proxmox-cli auth login -u <username>' to authenticate")
				return nil
			}

			// Test connection
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			if verbose {
				fmt.Fprintln(out, "\nTesting connection...")

				client, err := utility.AuthenticatedClient()
				if err != nil {
					return fmt.Errorf("connection failed: %w", err)
				}
				ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
				defer cancel()

				version, err := client.Version(ctx)
				if err != nil {
					fmt.Fprintln(out, "Try running 'proxmox-cli auth login -u <username>' to re-authenticate")
					return fmt.Errorf("connection failed: %w", err)
				}
				fmt.Fprintln(out, "Connection successful")
				fmt.Fprintf(out, "Proxmox VE Version: %s\n", version.Version)
				fmt.Fprintf(out, "Release: %s\n", version.Release)

				nodes, err := client.Nodes(ctx)
				if err != nil {
					return fmt.Errorf("fetch cluster nodes: %w", err)
				}
				fmt.Fprintf(out, "\nCluster nodes: %d\n", len(nodes))
				for _, node := range nodes {
					fmt.Fprintf(out, "   - %s %s\n", node.Node, node.Status)
				}
			} else {
				fmt.Fprintln(out, "\nUse --verbose to test the connection")
			}

			fmt.Fprintln(out, "\nAvailable commands:")
			fmt.Fprintln(out, "   - proxmox-cli nodes get     # List cluster nodes")
			fmt.Fprintln(out, "   - proxmox-cli vm get        # List virtual machines")
			fmt.Fprintln(out, "   - proxmox-cli lxc get       # List containers")
			fmt.Fprintln(out, "   - proxmox-cli --help        # Show all commands")
			return nil
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Test connection and show detailed information")
	return cmd
}
