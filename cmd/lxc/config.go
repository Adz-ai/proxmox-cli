package lxc

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
		Short: "Manage LXC container configuration",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newConfigSetCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set key=value [key=value...]",
		Short: "Set LXC container configuration options",
		Long: `Apply one or more configuration options to a container, e.g.:

  proxmox-cli lxc config set -n pve -i 200 memory=2048 swap=512`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()

			options := make([]proxmox.ContainerOption, 0, len(args))
			for _, arg := range args {
				key, value, found := strings.Cut(arg, "=")
				key = strings.TrimSpace(key)
				if !found || key == "" {
					return fmt.Errorf("invalid option %q; use key=value", arg)
				}
				options = append(options, proxmox.ContainerOption{Name: key, Value: value})
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.Config(ctx, options...)
			if err != nil {
				return fmt.Errorf("update config of container %d: %w", vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("update config of container %d: %w", vmid, err)
			}

			fmt.Fprintf(out, "Configuration of container %d updated successfully\n", vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	return cmd
}

func newResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Grow an LXC container volume",
		Long: `Increase the size of a container volume, e.g.:

  proxmox-cli lxc resize -n pve -i 200 --disk rootfs --size +2G`,
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

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := container.Resize(ctx, disk, size)
			if err != nil {
				return fmt.Errorf("resize volume %q of container %d: %w", disk, vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("resize volume %q of container %d: %w", disk, vmid, err)
			}

			fmt.Fprintf(out, "Volume %s of container %d resized successfully\n", disk, vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().String("disk", "", "Volume to resize, e.g. rootfs")
	cmd.Flags().String("size", "", "New size or increment, e.g. 16G or +2G")
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
		Short: "Add or remove LXC container tags",
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

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			for _, tag := range add {
				task, err := container.AddTag(ctx, tag)
				if err != nil {
					return fmt.Errorf("add tag %q to container %d: %w", tag, vmid, err)
				}
				if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
					return fmt.Errorf("add tag %q to container %d: %w", tag, vmid, err)
				}
			}
			for _, tag := range remove {
				task, err := container.RemoveTag(ctx, tag)
				if err != nil {
					return fmt.Errorf("remove tag %q from container %d: %w", tag, vmid, err)
				}
				if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
					return fmt.Errorf("remove tag %q from container %d: %w", tag, vmid, err)
				}
			}

			fmt.Fprintf(out, "Tags of container %d updated successfully\n", vmid)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	cmd.Flags().StringSlice("add", nil, "Tags to add (repeatable or comma-separated)")
	cmd.Flags().StringSlice("remove", nil, "Tags to remove (repeatable or comma-separated)")
	return cmd
}
