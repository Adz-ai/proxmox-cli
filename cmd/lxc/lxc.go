package lxc

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "lxc",
	Short: "Manage LXC containers",
	Long:  `Perform operations on LXC containers including create, delete, start, stop, and more.`,
}

func init() {
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(deleteCmd)
	// TODO: Implement these commands when we better understand the API
	// Cmd.AddCommand(statusCmd)
	// Cmd.AddCommand(describeCmd)
	// Cmd.AddCommand(restartCmd)
	// Cmd.AddCommand(cloneCmd)
	// Cmd.AddCommand(snapshotCmd)
}
