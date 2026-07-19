// Package backup implements vzdump backup creation, listing, and restore.
package backup

import (
	"fmt"
	"strings"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create, list, and restore guest backups",
		Long:  `Manage vzdump backups of VMs and LXC containers.`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newCreateCmd(), newListCmd(), newRestoreCmd())
	return cmd
}

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Back up a VM or container",
		Long: `Create a vzdump backup of a guest, e.g.:

  proxmox-cli backup create -n pve -i 100 --storage local --mode snapshot`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			vmid, err := cmd.Flags().GetInt("vmid")
			if err != nil {
				return fmt.Errorf("read vmid flag: %w", err)
			}
			storage, err := cmd.Flags().GetString("storage")
			if err != nil {
				return fmt.Errorf("read storage flag: %w", err)
			}
			mode, err := cmd.Flags().GetString("mode")
			if err != nil {
				return fmt.Errorf("read mode flag: %w", err)
			}
			compress, err := cmd.Flags().GetString("compress")
			if err != nil {
				return fmt.Errorf("read compress flag: %w", err)
			}
			notes, err := cmd.Flags().GetString("notes")
			if err != nil {
				return fmt.Errorf("read notes flag: %w", err)
			}
			if strings.TrimSpace(nodeName) == "" {
				return fmt.Errorf("node cannot be empty")
			}
			if vmid <= 0 {
				return fmt.Errorf("vmid must be positive")
			}
			switch mode {
			case "snapshot", "suspend", "stop":
			default:
				return fmt.Errorf("unsupported mode %q; use snapshot, suspend, or stop", mode)
			}
			switch compress {
			case "", "zstd", "gzip", "lzo":
			default:
				return fmt.Errorf("unsupported compression %q; use zstd, gzip, or lzo", compress)
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			options := &proxmox.VirtualMachineBackupOptions{
				VMID:          uint64(vmid),
				Storage:       strings.TrimSpace(storage),
				Mode:          proxmox.VirtualMachineBackupMode(mode),
				Compress:      proxmox.VirtualMachineBackupCompress(compress),
				NotesTemplate: notes,
			}
			task, err := node.Vzdump(ctx, options)
			if err != nil {
				return fmt.Errorf("back up guest %d: %w", vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("back up guest %d: %w", vmid, err)
			}

			fmt.Fprintf(out, "Backup of guest %d completed successfully\n", vmid)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node hosting the guest")
	cmd.Flags().IntP("vmid", "i", 0, "Guest ID to back up")
	cmd.Flags().String("storage", "", "Target storage for the backup")
	cmd.Flags().String("mode", "snapshot", "Backup mode: snapshot, suspend, or stop")
	cmd.Flags().String("compress", "zstd", "Compression: zstd, gzip, or lzo")
	cmd.Flags().String("notes", "", "Notes template for the backup")
	for _, flag := range []string{"node", "vmid"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	_ = cmd.RegisterFlagCompletionFunc("mode", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"snapshot", "suspend", "stop"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

type backupSummary struct {
	VolID     string `json:"volid"`
	VMID      uint64 `json:"vmid,omitempty"`
	Size      uint64 `json:"size_bytes"`
	Format    string `json:"format,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups on a storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			storageName, err := cmd.Flags().GetString("storage")
			if err != nil {
				return fmt.Errorf("read storage flag: %w", err)
			}
			vmid, err := cmd.Flags().GetInt("vmid")
			if err != nil {
				return fmt.Errorf("read vmid flag: %w", err)
			}
			if strings.TrimSpace(nodeName) == "" {
				return fmt.Errorf("node cannot be empty")
			}
			if strings.TrimSpace(storageName) == "" {
				return fmt.Errorf("storage cannot be empty")
			}
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			storage, err := node.Storage(ctx, storageName)
			if err != nil {
				return fmt.Errorf("get storage %q: %w", storageName, err)
			}

			content, err := storage.GetContent(ctx)
			if err != nil {
				return fmt.Errorf("list content of storage %q: %w", storageName, err)
			}

			summaries := []backupSummary{}
			for _, item := range content {
				if !strings.Contains(item.Volid, "backup/") {
					continue
				}
				if vmid > 0 && item.VMID != uint64(vmid) {
					continue
				}
				created := ""
				if item.Ctime > 0 {
					created = time.Unix(int64(item.Ctime), 0).UTC().Format(time.RFC3339)
				}
				summaries = append(summaries, backupSummary{
					VolID:     item.Volid,
					VMID:      item.VMID,
					Size:      item.Size,
					Format:    item.Format,
					CreatedAt: created,
					Notes:     item.Notes,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "Backups on %s:\n", storageName)
			fmt.Fprintln(out, "================")
			for _, summary := range summaries {
				created := summary.CreatedAt
				if created == "" {
					created = "N/A"
				}
				fmt.Fprintf(out, "%-70s %8.2f GiB  %-22s %s\n",
					summary.VolID, float64(summary.Size)/(1024*1024*1024), created, summary.Notes)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No backups found")
			}
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node to query")
	cmd.Flags().String("storage", "", "Storage holding the backups")
	cmd.Flags().IntP("vmid", "i", 0, "Only list backups of this guest")
	for _, flag := range []string{"node", "storage"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddOutputFlag(cmd)
	return cmd
}

func newRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a guest from a backup archive",
		Long: `Restore a VM or container from a vzdump archive, e.g.:

  proxmox-cli backup restore -n pve -i 100 --archive 'local:backup/vzdump-qemu-100-....vma.zst'

The guest type is detected from the archive name. Use --force to overwrite
an existing guest with the same ID.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			vmid, err := cmd.Flags().GetInt("vmid")
			if err != nil {
				return fmt.Errorf("read vmid flag: %w", err)
			}
			archive, err := cmd.Flags().GetString("archive")
			if err != nil {
				return fmt.Errorf("read archive flag: %w", err)
			}
			storage, err := cmd.Flags().GetString("storage")
			if err != nil {
				return fmt.Errorf("read storage flag: %w", err)
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("read force flag: %w", err)
			}
			if strings.TrimSpace(nodeName) == "" {
				return fmt.Errorf("node cannot be empty")
			}
			if vmid <= 0 {
				return fmt.Errorf("vmid must be positive")
			}
			archive = strings.TrimSpace(archive)
			if archive == "" {
				return fmt.Errorf("archive cannot be empty")
			}

			if force {
				if err := utility.ConfirmAction(cmd, fmt.Sprintf("Overwrite existing guest %d with archive %q?", vmid, archive)); err != nil {
					return err
				}
			}

			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			var task *proxmox.Task
			switch {
			case strings.Contains(archive, "vzdump-qemu-"):
				options := []proxmox.VirtualMachineOption{{Name: "archive", Value: archive}}
				if force {
					options = append(options, proxmox.VirtualMachineOption{Name: "force", Value: 1})
				}
				if storage != "" {
					options = append(options, proxmox.VirtualMachineOption{Name: "storage", Value: storage})
				}
				task, err = node.NewVirtualMachine(ctx, vmid, options...)
			case strings.Contains(archive, "vzdump-lxc-"):
				options := []proxmox.ContainerOption{
					{Name: "ostemplate", Value: archive},
					{Name: "restore", Value: 1},
				}
				if force {
					options = append(options, proxmox.ContainerOption{Name: "force", Value: 1})
				}
				if storage != "" {
					options = append(options, proxmox.ContainerOption{Name: "storage", Value: storage})
				}
				task, err = node.NewContainer(ctx, vmid, options...)
			default:
				return fmt.Errorf("cannot detect guest type from archive %q; expected a vzdump-qemu or vzdump-lxc archive", archive)
			}
			if err != nil {
				return fmt.Errorf("restore guest %d from %q: %w", vmid, archive, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("restore guest %d from %q: %w", vmid, archive, err)
			}

			fmt.Fprintf(out, "Guest %d restored successfully from %s\n", vmid, archive)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node to restore onto")
	cmd.Flags().IntP("vmid", "i", 0, "Guest ID to restore as")
	cmd.Flags().String("archive", "", "Backup volume ID to restore from")
	cmd.Flags().String("storage", "", "Target storage for restored disks")
	cmd.Flags().Bool("force", false, "Overwrite an existing guest with the same ID")
	for _, flag := range []string{"node", "vmid", "archive"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
	utility.AddYesFlag(cmd)
	return cmd
}
