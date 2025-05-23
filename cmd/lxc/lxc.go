package lxc

import (
	"github.com/spf13/cobra"
)

var LXCCmd = &cobra.Command{
	Use:   "lxc",
	Short: "Manage LXC containers",
	Long:  `Perform operations on LXC containers including create, delete, start, stop, and more.`,
}

func init() {
	LXCCmd.AddCommand(getCmd)
	// TODO: Implement these commands when we understand the API better
	// LXCCmd.AddCommand(createCmd)
	// LXCCmd.AddCommand(deleteCmd)
	// LXCCmd.AddCommand(describeCmd)
	// LXCCmd.AddCommand(startCmd)
	// LXCCmd.AddCommand(stopCmd)
	// LXCCmd.AddCommand(restartCmd)
	// LXCCmd.AddCommand(cloneCmd)
	// LXCCmd.AddCommand(snapshotCmd)
}