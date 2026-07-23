package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/luthermonson/go-proxmox"
	"go.uber.org/mock/gomock"

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
