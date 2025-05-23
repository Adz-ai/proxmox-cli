package nodes

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage nodes",
	Long:  "Manage nodes in the Proxmox cluster",
}

func init() {
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(describeCmd)
	// TODO: Add these commands when we understand the API better
	// Cmd.AddCommand(storageCmd)
	// Cmd.AddCommand(tasksCmd)
	// Cmd.AddCommand(servicesCmd)
}
