package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
		Server:      displayServer(utility.ContextString("server_url")),
		User:        displayUser(),
		CLIVersion:  cmd.Root().Version,
		Refresh:     refresh,
	})
}

// displayServer strips the scheme so the cluster info block stays compact.
func displayServer(serverURL string) string {
	serverURL = strings.TrimPrefix(serverURL, "https://")
	serverURL = strings.TrimPrefix(serverURL, "http://")
	return strings.TrimSuffix(serverURL, "/")
}

// displayUser is the authenticated identity: the API token ID when token
// auth is configured, otherwise the username embedded in the session ticket
// (formatted PVE:user@realm:...).
func displayUser() string {
	if tokenID := utility.ContextString("api_token.token_id"); tokenID != "" {
		return tokenID
	}
	ticket := utility.ContextString("auth_ticket.ticket")
	if parts := strings.Split(ticket, ":"); len(parts) >= 2 && parts[1] != "" {
		return parts[1]
	}
	return "n/a"
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

func (d *tuiDataSource) Version(ctx context.Context) (string, error) {
	version, err := d.client.Version(ctx)
	if err != nil {
		return "", fmt.Errorf("get PVE version: %w", err)
	}
	return version.Version, nil
}

// Tasks aggregates recent tasks from every node in the cluster.
func (d *tuiDataSource) Tasks(ctx context.Context) ([]tui.Resource, error) {
	statuses, err := d.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	rows := []tui.Resource{}
	for _, status := range statuses {
		node, err := d.client.Node(ctx, status.Node)
		if err != nil {
			return nil, fmt.Errorf("get node %q: %w", status.Node, err)
		}
		tasks, err := node.Tasks(ctx, &proxmox.NodeTasksOptions{Limit: 100, Source: "all"})
		if err != nil {
			return nil, fmt.Errorf("list tasks on node %q: %w", status.Node, err)
		}
		for _, task := range tasks {
			taskStatus := task.Status
			if task.IsRunning {
				taskStatus = "running"
			} else if task.ExitStatus != "" {
				taskStatus = task.ExitStatus
			}
			rows = append(rows, tui.Resource{
				Kind:   tui.KindTask,
				ID:     string(task.UPID),
				Name:   task.Type,
				Target: task.ID,
				Node:   task.Node,
				User:   task.User,
				Status: taskStatus,
				Start:  task.StartTime.Unix(),
				End:    task.EndTime.Unix(),
			})
		}
	}
	return rows, nil
}

// Snapshots lists a guest's snapshots, skipping the synthetic "current"
// entry Proxmox appends to represent the live state.
func (d *tuiDataSource) Snapshots(ctx context.Context, resource tui.Resource) ([]tui.Snapshot, error) {
	node, err := d.client.Node(ctx, resource.Node)
	if err != nil {
		return nil, fmt.Errorf("get node %q: %w", resource.Node, err)
	}
	items := []tui.Snapshot{}
	switch resource.Kind {
	case tui.KindVM:
		vm, err := node.VirtualMachine(ctx, int(resource.VMID))
		if err != nil {
			return nil, fmt.Errorf("get VM %d: %w", resource.VMID, err)
		}
		snapshots, err := vm.Snapshots(ctx)
		if err != nil {
			return nil, fmt.Errorf("list snapshots for VM %d: %w", resource.VMID, err)
		}
		for _, snapshot := range snapshots {
			if snapshot.Name == "current" {
				continue
			}
			items = append(items, tui.Snapshot{
				Name:        snapshot.Name,
				Parent:      snapshot.Parent,
				Description: snapshot.Description,
				Created:     snapshot.Snaptime,
			})
		}
	case tui.KindLXC:
		container, err := node.Container(ctx, int(resource.VMID))
		if err != nil {
			return nil, fmt.Errorf("get container %d: %w", resource.VMID, err)
		}
		snapshots, err := container.Snapshots(ctx)
		if err != nil {
			return nil, fmt.Errorf("list snapshots for container %d: %w", resource.VMID, err)
		}
		for _, snapshot := range snapshots {
			if snapshot.Name == "current" {
				continue
			}
			items = append(items, tui.Snapshot{
				Name:        snapshot.Name,
				Parent:      snapshot.Parent,
				Description: snapshot.Description,
				Created:     snapshot.SnapshotCreationTime,
			})
		}
	default:
		return nil, fmt.Errorf("snapshots are only supported for VMs and containers")
	}
	return items, nil
}

// Shell attaches an interactive console to a guest. Proxmox rejects
// websocket connections authenticated with API tokens, so this goes through
// a session-ticket-only client regardless of how the TUI itself connected.
func (d *tuiDataSource) Shell(resource tui.Resource) (tui.ShellSession, error) {
	if resource.Kind != tui.KindVM && resource.Kind != tui.KindLXC {
		return nil, errors.New("console is only available for VMs and containers")
	}
	// Re-read the config so a login done in another terminal while the TUI
	// is running is picked up; a failed read surfaces via SessionClient.
	_ = utility.LoadConfig()
	client, err := utility.SessionClient()
	if err != nil {
		return nil, err
	}
	return func(stdin io.Reader, stdout, stderr io.Writer) error {
		ctx := context.Background()
		node, err := client.Node(ctx, resource.Node)
		if err != nil {
			return fmt.Errorf("get node %q: %w", resource.Node, err)
		}
		var send, recv chan []byte
		var errs chan error
		var closer func() error
		switch resource.Kind {
		case tui.KindVM:
			vm, vmErr := node.VirtualMachine(ctx, int(resource.VMID))
			if vmErr != nil {
				return fmt.Errorf("get VM %d: %w", resource.VMID, vmErr)
			}
			term, termErr := vm.TermProxy(ctx)
			if termErr != nil {
				return fmt.Errorf("open terminal proxy: %w", termErr)
			}
			send, recv, errs, closer, err = vm.TermWebSocket(term)
		default:
			container, ctErr := node.Container(ctx, int(resource.VMID))
			if ctErr != nil {
				return fmt.Errorf("get container %d: %w", resource.VMID, ctErr)
			}
			term, termErr := container.TermProxy(ctx)
			if termErr != nil {
				return fmt.Errorf("open terminal proxy: %w", termErr)
			}
			send, recv, errs, closer, err = container.TermWebSocket(term)
		}
		if err != nil {
			return fmt.Errorf("open console websocket: %w", err)
		}
		return utility.RunConsoleStreams(ctx, stdin, stdout, send, recv, errs, closer)
	}, nil
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
	case tui.ActionDelete:
		return vm.Delete(ctx, nil)
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
	case tui.ActionDelete:
		return container.Delete(ctx, nil)
	default:
		return nil, fmt.Errorf("unsupported action %q", action)
	}
}
