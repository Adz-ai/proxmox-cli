package vm

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func newTargetFlagCmd(node, id string) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	addVMTargetFlags(cmd)
	if node != "" {
		_ = cmd.Flags().Set("node", node)
	}
	if id != "" {
		_ = cmd.Flags().Set("id", id)
	}
	return cmd
}

func TestVMTargetFromFlags(t *testing.T) {
	node, id, err := vmTargetFromFlags(newTargetFlagCmd(" pve ", "100"))
	if err != nil {
		t.Fatal(err)
	}
	if node != "pve" || id != 100 {
		t.Fatalf("got (%q, %d), want (\"pve\", 100)", node, id)
	}

	if _, _, err := vmTargetFromFlags(newTargetFlagCmd("  ", "100")); err == nil {
		t.Error("expected error for blank node")
	}
	if _, _, err := vmTargetFromFlags(newTargetFlagCmd("pve", "0")); err == nil {
		t.Error("expected error for nonpositive id")
	}
}

func TestMapToVMOptions(t *testing.T) {
	if _, err := mapToVMOptions(map[string]interface{}{}); err == nil {
		t.Error("expected error for empty spec")
	}
	if _, err := mapToVMOptions(map[string]interface{}{"VMID": 100}); err == nil {
		t.Error("expected error when spec overrides vmid")
	}
	if _, err := mapToVMOptions(map[string]interface{}{"node": "pve"}); err == nil {
		t.Error("expected error when spec overrides node")
	}

	options, err := mapToVMOptions(map[string]interface{}{"memory": 2048, "cores": 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 2 || options[0].Name != "cores" || options[1].Name != "memory" {
		t.Fatalf("options not sorted by name: %+v", options)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(1024 * 1024 * 1024); got != "1.00 GiB" {
		t.Errorf("formatBytes(1 GiB) = %q", got)
	}
}

func TestFormatUptime(t *testing.T) {
	if got := formatUptime(90061); !strings.Contains(got, "1 days, 1 hours, 1 minutes, 1 seconds") {
		t.Errorf("formatUptime(90061) = %q", got)
	}
}

func setupVMMocks(t *testing.T) (*gomock.Controller, *mocks.MockProxmoxClientInterface) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")

	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface { return client })
	t.Cleanup(utility.ResetClientFactory)
	return ctrl, client
}

func TestStopCommand(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().Stop(ctx).Return(&proxmox.Task{IsSuccessful: true}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"stop", "-n", "pve", "-i", "100"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "VM 100 stopped successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestSnapshotRollbackCommand(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().RollbackSnapshot(ctx, "nightly").Return(&proxmox.Task{IsSuccessful: true}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"snapshot", "rollback", "-n", "pve", "-i", "100", "--name", "nightly"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `VM 100 rolled back to snapshot "nightly" successfully`) {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestVMCloneCommand(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().Clone(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, options *proxmox.VirtualMachineCloneOptions) (int, *proxmox.Task, error) {
			if options.NewID != 101 || options.Name != "copy" || !bool(options.Full) {
				t.Errorf("unexpected clone options: %+v", options)
			}
			return 101, &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"clone", "-n", "pve", "-s", "100", "-t", "101", "--name", "copy", "--full"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "VM 100 cloned to 101 successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestMigratePreflightBlocksRunningVMWithoutOnline(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().MigratePreconditions(ctx, "pve2").Return(
		&proxmox.VirtualMachineMigratePreconditions{Running: true}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"migrate", "-n", "pve", "-i", "100", "--target", "pve2"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "use --online") {
		t.Fatalf("expected preflight error, got %v", err)
	}
}

func TestMigrateCommand(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().MigratePreconditions(ctx, "pve2").Return(
		&proxmox.VirtualMachineMigratePreconditions{Running: true}, nil)
	vm.EXPECT().Migrate(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, options *proxmox.VirtualMachineMigrateOptions) (*proxmox.Task, error) {
			if options.Target != "pve2" || !bool(options.Online) {
				t.Errorf("unexpected migrate options: %+v", options)
			}
			return &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"migrate", "-n", "pve", "-i", "100", "--target", "pve2", "--online"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "VM 100 migrated to pve2 successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestConfigSetCommand(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().Config(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, options ...proxmox.VirtualMachineOption) (*proxmox.Task, error) {
			values := map[string]any{}
			for _, option := range options {
				values[option.Name] = option.Value
			}
			if values["memory"] != "4096" || values["cores"] != "4" {
				t.Errorf("unexpected config options: %v", values)
			}
			return &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"config", "set", "-n", "pve", "-i", "100", "memory=4096", "cores=4"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Configuration of VM 100 updated successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestCreateAutoAssignsID(t *testing.T) {
	ctrl, client := setupVMMocks(t)
	node := mocks.NewMockNodeInterface(ctrl)
	cluster := mocks.NewMockClusterInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Cluster(ctx).Return(cluster, nil)
	cluster.EXPECT().NextID(ctx).Return(105, nil)
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().NewVirtualMachine(ctx, 105, gomock.Any()).Return(&proxmox.Task{IsSuccessful: true}, nil)

	spec := filepath.Join(t.TempDir(), "vm.yaml")
	if err := os.WriteFile(spec, []byte("memory: 2048\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"create", "-n", "pve", "-s", spec})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Virtual machine 105 created successfully") {
		t.Fatalf("expected auto-assigned ID in output:\n%s", out.String())
	}
}
