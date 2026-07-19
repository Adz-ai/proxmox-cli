package nodes

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Manage nodes",
		Long:  "Manage nodes in the Proxmox cluster",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newDescribeCmd())
	cmd.AddCommand(newStorageCmd())
	cmd.AddCommand(newTasksCmd())

	return cmd
}
