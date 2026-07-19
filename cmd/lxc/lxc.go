package lxc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lxc",
		Short: "Manage LXC containers",
		Long:  `Perform operations on LXC containers including create, delete, start, stop, and more.`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newGetCmd(),
		newCreateCmd(),
		newStartCmd(),
		newStopCmd(),
		newDeleteCmd(),
	)
	return cmd
}

func validateContainerTarget(node string, vmid int) error {
	if strings.TrimSpace(node) == "" {
		return fmt.Errorf("node cannot be empty")
	}
	if vmid <= 0 {
		return fmt.Errorf("container ID must be positive")
	}
	return nil
}
