package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/tui"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func TestTUIRequiresInteractiveTerminal(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupResourcesMocks(t, ctrl)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetIn(&bytes.Buffer{})
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"tui"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("expected interactive-terminal error, got %v", err)
	}
}

func TestRootTUIFlagLaunchesTUI(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupResourcesMocks(t, ctrl)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetIn(&bytes.Buffer{})
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--tui"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("expected the --tui flag to reach the TUI launcher, got %v", err)
	}
}

func TestTUICommandRejectsShortRefresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupResourcesMocks(t, ctrl)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"tui", "--refresh", "100ms"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "at least 1s") {
		t.Fatalf("expected refresh validation error, got %v", err)
	}
}

func TestTUIDataSourceMapsClusterResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	cluster := mocks.NewMockClusterInterface(ctrl)
	client.EXPECT().Cluster(gomock.Any()).Return(cluster, nil)
	cluster.EXPECT().Resources(gomock.Any()).Return(proxmox.ClusterResources{
		&proxmox.ClusterResource{ID: "qemu/100", Type: "qemu", VMID: 100, Name: "web", Node: "pve1", Status: "running", Mem: 512, MaxMem: 1024},
		&proxmox.ClusterResource{ID: "qemu/900", Type: "qemu", VMID: 900, Name: "tmpl", Node: "pve1", Status: "stopped", Template: 1},
		&proxmox.ClusterResource{ID: "lxc/200", Type: "lxc", VMID: 200, Name: "db", Node: "pve2", Status: "stopped"},
		&proxmox.ClusterResource{ID: "node/pve1", Type: "node", Node: "pve1", Status: "online", MaxCPU: 8},
		&proxmox.ClusterResource{ID: "storage/pve1/local", Type: "storage", Storage: "local", Node: "pve1", Status: "available", Shared: 1, PluginType: "dir"},
		&proxmox.ClusterResource{ID: "sdn/pve1/vnet0", Type: "sdn", Node: "pve1"},
	}, nil)

	source := &tuiDataSource{client: client}
	rows, err := source.Resources(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("expected unsupported types to be skipped, got %d rows", len(rows))
	}
	byID := map[string]tui.Resource{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	if byID["qemu/100"].Kind != tui.KindVM || byID["lxc/200"].Kind != tui.KindLXC {
		t.Fatalf("guest kinds mapped incorrectly: %+v", rows)
	}
	if !byID["qemu/900"].Template {
		t.Fatal("template flag should map to bool")
	}
	if byID["node/pve1"].Name != "pve1" {
		t.Fatalf("node rows should use the node name, got %q", byID["node/pve1"].Name)
	}
	storage := byID["storage/pve1/local"]
	if storage.Name != "local" || !storage.Shared || storage.Plugin != "dir" {
		t.Fatalf("storage row mapped incorrectly: %+v", storage)
	}
}

func TestTUIDataSourceGuestActions(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)

	client.EXPECT().Node(gomock.Any(), "pve1").Return(node, nil).Times(2)
	node.EXPECT().VirtualMachine(gomock.Any(), 100).Return(vm, nil)
	vm.EXPECT().Stop(gomock.Any()).Return(&proxmox.Task{}, nil)
	node.EXPECT().Container(gomock.Any(), 200).Return(container, nil)
	container.EXPECT().Shutdown(gomock.Any(), false, 600).Return(&proxmox.Task{}, nil)

	source := &tuiDataSource{client: client, timeout: 10 * time.Minute}
	if err := source.Guest(context.Background(), tui.Resource{Kind: tui.KindVM, VMID: 100, Node: "pve1"}, tui.ActionStop); err != nil {
		t.Fatalf("VM stop: %v", err)
	}
	if err := source.Guest(context.Background(), tui.Resource{Kind: tui.KindLXC, VMID: 200, Node: "pve1"}, tui.ActionShutdown); err != nil {
		t.Fatalf("container shutdown: %v", err)
	}
}

func TestTUIDataSourceVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	client.EXPECT().Version(gomock.Any()).Return(&proxmox.Version{Version: "8.4.1"}, nil)

	source := &tuiDataSource{client: client}
	version, err := source.Version(context.Background())
	if err != nil || version != "8.4.1" {
		t.Fatalf("version = %q, err = %v", version, err)
	}
}

func TestDisplayServerAndUser(t *testing.T) {
	if got := displayServer("https://pve.example.com:8006/"); got != "pve.example.com:8006" {
		t.Errorf("displayServer = %q", got)
	}
}

func TestTUIDataSourceTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	client.EXPECT().Nodes(gomock.Any()).Return(proxmox.NodeStatuses{&proxmox.NodeStatus{Node: "pve1"}}, nil)
	client.EXPECT().Node(gomock.Any(), "pve1").Return(node, nil)
	node.EXPECT().Tasks(gomock.Any(), &proxmox.NodeTasksOptions{Limit: 100, Source: "all"}).Return([]*proxmox.Task{
		{UPID: "UPID:pve1:1", ID: "110", Type: "vzdump", User: "root@pam", Node: "pve1", ExitStatus: "OK", StartTime: time.Unix(100, 0), EndTime: time.Unix(160, 0)},
		{UPID: "UPID:pve1:2", ID: "109", Type: "qmstart", User: "root@pam", Node: "pve1", IsRunning: true, StartTime: time.Unix(200, 0)},
	}, nil)

	source := &tuiDataSource{client: client}
	rows, err := source.Tasks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(rows))
	}
	if rows[0].Kind != tui.KindTask || rows[0].Status != "OK" || rows[0].Target != "110" || rows[0].Start != 100 {
		t.Fatalf("completed task mapped incorrectly: %+v", rows[0])
	}
	if rows[1].Status != "running" {
		t.Fatalf("running task should report running, got %+v", rows[1])
	}
}

func TestTUIDataSourceSnapshotsSkipCurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)
	client.EXPECT().Node(gomock.Any(), "pve1").Return(node, nil)
	node.EXPECT().VirtualMachine(gomock.Any(), 100).Return(vm, nil)
	vm.EXPECT().Snapshots(gomock.Any()).Return([]*proxmox.VirtualMachineSnapshot{
		{Name: "pre-upgrade", Snaptime: 1700000000, Description: "before v2"},
		{Name: "current", Description: "You are here!"},
	}, nil)

	source := &tuiDataSource{client: client}
	items, err := source.Snapshots(context.Background(), tui.Resource{Kind: tui.KindVM, VMID: 100, Node: "pve1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "pre-upgrade" || items[0].Created != 1700000000 {
		t.Fatalf("snapshots mapped incorrectly: %+v", items)
	}
}

func TestTUIDataSourceShellRequiresSessionTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupResourcesMocks(t, ctrl)

	// Shell reloads the config from disk, so persist each credential state.
	viper.Set("auth_ticket.ticket", "")
	viper.Set("auth_ticket.CSRFPreventionToken", "")
	viper.Set("api_token.token_id", "root@pam!cli")
	viper.Set("api_token.secret", "secret")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	source := &tuiDataSource{client: mocks.NewMockProxmoxClientInterface(ctrl)}
	guest := tui.Resource{Kind: tui.KindVM, VMID: 100, Node: "pve1"}
	if _, err := source.Shell(guest); err == nil || !strings.Contains(err.Error(), "API tokens cannot open websockets") {
		t.Fatalf("token-only auth should refuse the console, got %v", err)
	}

	viper.Set("auth_ticket.ticket", "PVE:root@pam:aa")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}
	session, err := source.Shell(guest)
	if err != nil || session == nil {
		t.Fatalf("session auth should return a console session, got %v", err)
	}

	if _, err := source.Shell(tui.Resource{Kind: tui.KindNode, Node: "pve1"}); err == nil {
		t.Fatal("console must be refused for non-guests")
	}
}

func TestTUIDataSourceRejectsNonGuests(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	client.EXPECT().Node(gomock.Any(), "pve1").Return(node, nil)

	source := &tuiDataSource{client: client, timeout: time.Minute}
	err := source.Guest(context.Background(), tui.Resource{Kind: tui.KindNode, Node: "pve1"}, tui.ActionStart)
	if err == nil || !strings.Contains(err.Error(), "only supported") {
		t.Fatalf("expected non-guest rejection, got %v", err)
	}
}
