package cmd

import (
	"github.com/Adz-ai/proxmox-cli/cmd/auth"
	"github.com/Adz-ai/proxmox-cli/cmd/backup"
	"github.com/Adz-ai/proxmox-cli/cmd/lxc"
	"github.com/Adz-ai/proxmox-cli/cmd/nodes"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/cmd/vm"

	"github.com/spf13/cobra"
)

// NewRootCmd creates a new root command for testing
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "proxmox-cli",
		Short:        "Command-line interface for Proxmox VE",
		SilenceUsage: true,
		Long: `Proxmox CLI - Manage your Proxmox Virtual Environment from the terminal

A command-line tool for managing Proxmox VE resources including:
- Virtual Machines (VMs)
- LXC Containers
- Cluster Nodes
- Storage and Networks

Get started:
  proxmox-cli init                    # Configure connection
  proxmox-cli auth login -u root@pam  # Authenticate
  proxmox-cli status                  # Check connection`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return utility.LoadConfig()
		},
	}

	cmd.PersistentFlags().Duration("timeout", utility.DefaultTaskTimeout, "Maximum time to wait for a Proxmox task to complete")

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newResourcesCmd())
	cmd.AddCommand(nodes.NewCmd())
	cmd.AddCommand(auth.NewCmd())
	cmd.AddCommand(vm.NewCmd())
	cmd.AddCommand(lxc.NewCmd())
	cmd.AddCommand(backup.NewCmd())

	return cmd
}
