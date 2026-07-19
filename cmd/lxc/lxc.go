package lxc

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
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
		newDescribeCmd(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newShutdownCmd(),
		newSuspendCmd(),
		newResumeCmd(),
		newDeleteCmd(),
		newCloneCmd(),
		newSnapshotCmd(),
		newMigrateCmd(),
		newConfigCmd(),
		newResizeCmd(),
		newTagsCmd(),
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

func addContainerTargetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("vmid", "i", 0, "Container ID")
	for _, flag := range []string{"node", "vmid"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
}

func containerTargetFromFlags(cmd *cobra.Command) (string, int, error) {
	node, err := cmd.Flags().GetString("node")
	if err != nil {
		return "", 0, fmt.Errorf("read node flag: %w", err)
	}
	vmid, err := cmd.Flags().GetInt("vmid")
	if err != nil {
		return "", 0, fmt.Errorf("read vmid flag: %w", err)
	}
	if err := validateContainerTarget(node, vmid); err != nil {
		return "", 0, err
	}
	return node, vmid, nil
}
