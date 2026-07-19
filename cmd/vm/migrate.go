package vm

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
		Short: "Migrate a virtual machine to another node",
		Long: `Move a virtual machine to a different cluster node. Preconditions are
checked first so blocking problems (local resources, disallowed targets)
are reported before the migration starts.`,
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
			online, err := cmd.Flags().GetBool("online")
			if err != nil {
				return fmt.Errorf("read online flag: %w", err)
			}
			withLocalDisks, err := cmd.Flags().GetBool("with-local-disks")
			if err != nil {
				return fmt.Errorf("read with-local-disks flag: %w", err)
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			preconditions, err := vm.MigratePreconditions(ctx, target)
			if err != nil {
				return fmt.Errorf("check migration preconditions for VM %d: %w", id, err)
			}
			if blocked, exists := preconditions.NotAllowedNodes[target]; exists && blocked != nil {
				return fmt.Errorf("target node %q is not allowed for VM %d", target, id)
			}
			if len(preconditions.LocalResources) > 0 {
				return fmt.Errorf("VM %d uses local resources that block migration: %s",
					id, strings.Join(preconditions.LocalResources, ", "))
			}
			if preconditions.Running && !online {
				return fmt.Errorf("VM %d is running; use --online for live migration or shut it down first", id)
			}
			if len(preconditions.LocalDisks) > 0 && !withLocalDisks {
				return fmt.Errorf("VM %d has local disks; use --with-local-disks to migrate them", id)
			}

			task, err := vm.Migrate(ctx, &proxmox.VirtualMachineMigrateOptions{
				Target:         target,
				Online:         proxmox.IntOrBool(online),
				WithLocalDisks: proxmox.IntOrBool(withLocalDisks),
			})
			if err != nil {
				return fmt.Errorf("migrate VM %d to %q: %w", id, target, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("migrate VM %d to %q: %w", id, target, err)
			}

			fmt.Fprintf(out, "VM %d migrated to %s successfully\n", id, target)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	cmd.Flags().String("target", "", "Destination node")
	cmd.Flags().Bool("online", false, "Live-migrate a running VM")
	cmd.Flags().Bool("with-local-disks", false, "Also migrate local disks")
	if err := cmd.MarkFlagRequired("target"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "target")
	return cmd
}
