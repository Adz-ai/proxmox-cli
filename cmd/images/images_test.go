package images

import (
	"bytes"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func setupImagesMocks(t *testing.T, ctrl *gomock.Controller) *mocks.MockProxmoxClientInterface {
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

func TestTemplateAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupImagesMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Appliances(ctx).Return(proxmox.Appliances{
		&proxmox.Appliance{Template: "debian-12-standard_12.7-1_amd64.tar.zst", Os: "debian-12", Description: "Debian 12 standard"},
	}, nil)

	cmd := NewTemplateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"available", "-n", "pve"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "debian-12-standard") {
		t.Fatalf("expected template in output:\n%s", out.String())
	}
}

func TestTemplateDownloadStartsTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupImagesMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().DownloadAppliance(ctx, "debian-12-standard_12.7-1_amd64.tar.zst", "local").
		Return("UPID:pve:00003333:00112233:65432100:download::root@pam:", nil)

	cmd := NewTemplateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"download", "-n", "pve", "--storage", "local",
		"--template", "debian-12-standard_12.7-1_amd64.tar.zst"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Download of debian-12-standard_12.7-1_amd64.tar.zst started") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestISOListFiltersToISOs(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupImagesMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	storage := mocks.NewMockStorageInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Storage(ctx, "local").Return(storage, nil)
	storage.EXPECT().GetContent(ctx).Return([]*proxmox.StorageContent{
		{Volid: "local:iso/debian-12.iso", Size: 700 * 1024 * 1024},
		{Volid: "local:backup/vzdump-qemu-100.vma.zst", Size: 1024},
	}, nil)

	cmd := NewISOCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "-n", "pve", "--storage", "local"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "debian-12.iso") {
		t.Fatalf("expected ISO in output:\n%s", out.String())
	}
	if strings.Contains(out.String(), "vzdump") {
		t.Fatalf("backups should be filtered out:\n%s", out.String())
	}
}
