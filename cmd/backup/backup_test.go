package backup

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func setupBackupMocks(t *testing.T, ctrl *gomock.Controller) *mocks.MockProxmoxClientInterface {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")

	client := mocks.NewMockProxmoxClientInterface(ctrl)
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface { return client })
	t.Cleanup(utility.ResetClientFactory)
	return client
}

func TestBackupCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupBackupMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Vzdump(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, options *proxmox.VirtualMachineBackupOptions) (*proxmox.Task, error) {
			if options.VMID != 100 || options.Storage != "local" || string(options.Mode) != "snapshot" {
				t.Errorf("unexpected vzdump options: %+v", options)
			}
			return &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"create", "-n", "pve", "-i", "100", "--storage", "local"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Backup of guest 100 completed successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestBackupCreateRejectsBadMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupBackupMocks(t, ctrl)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"create", "-n", "pve", "-i", "100", "--mode", "live"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("expected unsupported-mode error, got %v", err)
	}
}

func TestBackupListFiltersToBackups(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupBackupMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	storage := mocks.NewMockStorageInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Storage(ctx, "local").Return(storage, nil)
	storage.EXPECT().GetContent(ctx).Return([]*proxmox.StorageContent{
		{Volid: "local:backup/vzdump-qemu-100-2026_07_19-10_00_00.vma.zst", VMID: 100, Size: 2 * 1024 * 1024 * 1024, Ctime: 1780000000},
		{Volid: "local:backup/vzdump-lxc-200-2026_07_18-01_00_00.tar.zst", VMID: 200, Size: 1024 * 1024 * 1024},
		{Volid: "local:iso/debian-12.iso", Size: 700 * 1024 * 1024},
	}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "-n", "pve", "--storage", "local"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "vzdump-qemu-100") || !strings.Contains(out.String(), "vzdump-lxc-200") {
		t.Fatalf("expected both backups in output:\n%s", out.String())
	}
	if strings.Contains(out.String(), "debian-12.iso") {
		t.Fatalf("non-backup content should be filtered out:\n%s", out.String())
	}
}

func TestBackupRestoreDetectsGuestType(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupBackupMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().NewContainer(ctx, 200, gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, vmid int, options ...proxmox.ContainerOption) (*proxmox.Task, error) {
			values := map[string]any{}
			for _, option := range options {
				values[option.Name] = option.Value
			}
			if values["restore"] != 1 || values["force"] != 1 {
				t.Errorf("unexpected restore options: %v", values)
			}
			return &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"restore", "-n", "pve", "-i", "200",
		"--archive", "local:backup/vzdump-lxc-200-2026_07_18-01_00_00.tar.zst", "--force", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Guest 200 restored successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestBackupRestoreRejectsUnknownArchive(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupBackupMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	client.EXPECT().Node(gomock.Any(), "pve").Return(node, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"restore", "-n", "pve", "-i", "300", "--archive", "local:iso/debian.iso"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "cannot detect guest type") {
		t.Fatalf("expected detection error, got %v", err)
	}
}
