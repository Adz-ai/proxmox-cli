package vm

import (
	"bytes"
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

func TestStopCommand(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")

	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	vm := mocks.NewMockVirtualMachineInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().VirtualMachine(ctx, 100).Return(vm, nil)
	vm.EXPECT().Stop(ctx).Return(&proxmox.Task{IsSuccessful: true}, nil)

	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface { return client })
	t.Cleanup(utility.ResetClientFactory)

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
