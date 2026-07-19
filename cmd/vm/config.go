package vm

import (
	"fmt"
	"strings"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage virtual machine configuration",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newConfigSetCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set key=value [key=value...]",
		Short: "Set virtual machine configuration options",
		Long: `Apply one or more configuration options to a virtual machine, e.g.:

  proxmox-cli vm config set -n pve -i 100 memory=4096 cores=4`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()

			options := make([]proxmox.VirtualMachineOption, 0, len(args))
			for _, arg := range args {
				key, value, found := strings.Cut(arg, "=")
				key = strings.TrimSpace(key)
				if !found || key == "" {
					return fmt.Errorf("invalid option %q; use key=value", arg)
				}
				options = append(options, proxmox.VirtualMachineOption{Name: key, Value: value})
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := vm.Config(ctx, options...)
			if err != nil {
				return fmt.Errorf("update config of VM %d: %w", id, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("update config of VM %d: %w", id, err)
			}

			fmt.Fprintf(out, "Configuration of VM %d updated successfully\n", id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	return cmd
}

func newResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Grow a virtual machine disk",
		Long: `Increase the size of a VM disk, e.g.:

  proxmox-cli vm resize -n pve -i 100 --disk scsi0 --size +10G`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			disk, err := cmd.Flags().GetString("disk")
			if err != nil {
				return fmt.Errorf("read disk flag: %w", err)
			}
			size, err := cmd.Flags().GetString("size")
			if err != nil {
				return fmt.Errorf("read size flag: %w", err)
			}
			disk = strings.TrimSpace(disk)
			size = strings.TrimSpace(size)
			if disk == "" {
				return fmt.Errorf("disk cannot be empty")
			}
			if size == "" {
				return fmt.Errorf("size cannot be empty")
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := vm.ResizeDisk(ctx, disk, size)
			if err != nil {
				return fmt.Errorf("resize disk %q of VM %d: %w", disk, id, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("resize disk %q of VM %d: %w", disk, id, err)
			}

			fmt.Fprintf(out, "Disk %s of VM %d resized successfully\n", disk, id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	cmd.Flags().String("disk", "", "Disk to resize, e.g. scsi0")
	cmd.Flags().String("size", "", "New size or increment, e.g. 32G or +10G")
	for _, flag := range []string{"disk", "size"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	return cmd
}

func newTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Add or remove virtual machine tags",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			add, err := cmd.Flags().GetStringSlice("add")
			if err != nil {
				return fmt.Errorf("read add flag: %w", err)
			}
			remove, err := cmd.Flags().GetStringSlice("remove")
			if err != nil {
				return fmt.Errorf("read remove flag: %w", err)
			}
			if len(add) == 0 && len(remove) == 0 {
				return fmt.Errorf("nothing to do; use --add and/or --remove")
			}

			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			for _, tag := range add {
				task, err := vm.AddTag(ctx, tag)
				if err != nil {
					return fmt.Errorf("add tag %q to VM %d: %w", tag, id, err)
				}
				if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
					return fmt.Errorf("add tag %q to VM %d: %w", tag, id, err)
				}
			}
			for _, tag := range remove {
				task, err := vm.RemoveTag(ctx, tag)
				if err != nil {
					return fmt.Errorf("remove tag %q from VM %d: %w", tag, id, err)
				}
				if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
					return fmt.Errorf("remove tag %q from VM %d: %w", tag, id, err)
				}
			}

			fmt.Fprintf(out, "Tags of VM %d updated successfully\n", id)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	cmd.Flags().StringSlice("add", nil, "Tags to add (repeatable or comma-separated)")
	cmd.Flags().StringSlice("remove", nil, "Tags to remove (repeatable or comma-separated)")
	return cmd
}
