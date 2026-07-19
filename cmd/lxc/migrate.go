package lxc

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate an LXC container to another node",
		Long: `Move an LXC container to a different cluster node. Running containers
need --restart (stop, move, start) since LXC does not support live migration.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			target, err := cmd.Flags().GetString("target")
			if err != nil {
				return fmt.Errorf("read target flag: %w", err)
			}
			target = strings.TrimSpace(target)
			if target == "" {
				return fmt.Errorf("target node cannot be empty")
			}
			restart, err := cmd.Flags().GetBool("restart")
			if err != nil {
				return fmt.Errorf("read restart flag: %w", err)
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.Migrate(ctx, &proxmox.ContainerMigrateOptions{
				Target:  target,
				Restart: proxmox.IntOrBool(restart),
			})
			if err != nil {
				return fmt.Errorf("migrate container %d to %q: %w", vmid, target, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("migrate container %d to %q: %w", vmid, target, err)
			}

			fmt.Fprintf(out, "Container %d migrated to %s successfully\n", vmid, target)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().String("target", "", "Destination node")
	cmd.Flags().Bool("restart", false, "Restart-migrate a running container")
	if err := cmd.MarkFlagRequired("target"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "target")
	return cmd
}
