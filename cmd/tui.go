package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/internal/tui"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultTUIRefresh = 5 * time.Second

func newTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Browse and manage the cluster in an interactive terminal UI",
		Long: `Open a k9s-style terminal UI with live-refreshing views of guests,
nodes, and storage. Navigate with the keyboard, filter with /, and run
lifecycle actions (start, shutdown, stop, reboot) on the selected guest.
Destructive actions ask for confirmation inside the UI.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			refresh, err := cmd.Flags().GetDuration("refresh")
			if err != nil {
				return fmt.Errorf("get refresh flag: %w", err)
			}
			return launchTUI(cmd, refresh)
		},
	}

	cmd.Flags().Duration("refresh", defaultTUIRefresh, "Auto-refresh interval for cluster state")
	return cmd
}

func launchTUI(cmd *cobra.Command, refresh time.Duration) error {
	if refresh < time.Second {
		return errors.New("refresh interval must be at least 1s")
	}
	stdin, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return errors.New("the TUI requires an interactive terminal")
	}
	client, err := utility.AuthenticatedClient()
	if err != nil {
		return fmt.Errorf("authenticate Proxmox client: %w", err)
	}
	return tui.Run(tui.Options{
		Source:      &tuiDataSource{client: client, timeout: utility.TaskTimeout(cmd)},
		ContextName: utility.ActiveContext(),
		Refresh:     refresh,
	})
}

// tuiDataSource adapts the Proxmox client interfaces to the TUI's DataSource.
type tuiDataSource struct {
	client  interfaces.ProxmoxClientInterface
	timeout time.Duration
}

func (d *tuiDataSource) Resources(ctx context.Context) ([]tui.Resource, error) {
	cluster, err := d.client.Cluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cluster: %w", err)
	}
	resources, err := cluster.Resources(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cluster resources: %w", err)
	}
	rows := make([]tui.Resource, 0, len(resources))
	for _, resource := range resources {
		row := tui.Resource{
			ID:       resource.ID,
			VMID:     resource.VMID,
			Name:     resource.Name,
			Node:     resource.Node,
			Status:   resource.Status,
			CPU:      resource.CPU,
			MaxCPU:   resource.MaxCPU,
			Mem:      resource.Mem,
			MaxMem:   resource.MaxMem,
			Disk:     resource.Disk,
			MaxDisk:  resource.MaxDisk,
			Uptime:   resource.Uptime,
			Tags:     resource.Tags,
			Template: resource.Template == 1,
			HAState:  resource.HAstate,
			Shared:   resource.Shared == 1,
			Plugin:   resource.PluginType,
		}
		switch resource.Type {
		case "qemu":
			row.Kind = tui.KindVM
		case "lxc":
			row.Kind = tui.KindLXC
		case "node":
			row.Kind = tui.KindNode
			row.Name = resource.Node
		case "storage":
			row.Kind = tui.KindStorage
			row.Name = resource.Storage
		default:
			continue
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (d *tuiDataSource) Guest(ctx context.Context, resource tui.Resource, action tui.Action) error {
	node, err := d.client.Node(ctx, resource.Node)
	if err != nil {
		return fmt.Errorf("get node %q: %w", resource.Node, err)
	}
	var task *proxmox.Task
	switch resource.Kind {
	case tui.KindVM:
		vm, vmErr := node.VirtualMachine(ctx, int(resource.VMID))
		if vmErr != nil {
			return fmt.Errorf("get VM %d: %w", resource.VMID, vmErr)
		}
		task, err = vmTaskForAction(ctx, vm, action)
	case tui.KindLXC:
		container, ctErr := node.Container(ctx, int(resource.VMID))
		if ctErr != nil {
			return fmt.Errorf("get container %d: %w", resource.VMID, ctErr)
		}
		task, err = containerTaskForAction(ctx, container, action, d.timeout)
	default:
		return fmt.Errorf("actions are only supported for VMs and containers")
	}
	if err != nil {
		return fmt.Errorf("%s %s %d: %w", action, resource.Kind, resource.VMID, err)
	}
	return utility.WaitForTask(ctx, task, d.timeout, nil)
}

func vmTaskForAction(ctx context.Context, vm interfaces.VirtualMachineInterface, action tui.Action) (*proxmox.Task, error) {
	switch action {
	case tui.ActionStart:
		return vm.Start(ctx)
	case tui.ActionShutdown:
		return vm.Shutdown(ctx)
	case tui.ActionStop:
		return vm.Stop(ctx)
	case tui.ActionReboot:
		return vm.Reboot(ctx)
	default:
		return nil, fmt.Errorf("unsupported action %q", action)
	}
}

func containerTaskForAction(ctx context.Context, container interfaces.ContainerInterface, action tui.Action, timeout time.Duration) (*proxmox.Task, error) {
	switch action {
	case tui.ActionStart:
		return container.Start(ctx)
	case tui.ActionShutdown:
		return container.Shutdown(ctx, false, int(timeout.Seconds()))
	case tui.ActionStop:
		return container.Stop(ctx)
	case tui.ActionReboot:
		return container.Reboot(ctx)
	default:
		return nil, fmt.Errorf("unsupported action %q", action)
	}
}
