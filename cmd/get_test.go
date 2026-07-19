package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func setupResourcesMocks(t *testing.T, ctrl *gomock.Controller) *mocks.MockProxmoxClientInterface {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("PROXMOX_CLI_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	client := mocks.NewMockProxmoxClientInterface(ctrl)
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface { return client })
	t.Cleanup(utility.ResetClientFactory)
	return client
}

func clusterResourcesFixture() proxmox.ClusterResources {
	return proxmox.ClusterResources{
		&proxmox.ClusterResource{ID: "qemu/100", Type: "qemu", VMID: 100, Name: "web", Node: "pve1", Status: "running", Uptime: 86400},
		&proxmox.ClusterResource{ID: "lxc/200", Type: "lxc", VMID: 200, Name: "db", Node: "pve2", Status: "stopped"},
		&proxmox.ClusterResource{ID: "storage/pve1/local", Type: "storage", Storage: "local", Node: "pve1", Status: "available"},
		&proxmox.ClusterResource{ID: "node/pve1", Type: "node", Node: "pve1", Status: "online"},
	}
}

func TestResourcesCommandListsGuestsAndStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupResourcesMocks(t, ctrl)
	cluster := mocks.NewMockClusterInterface(ctrl)
	client.EXPECT().Cluster(gomock.Any()).Return(cluster, nil)
	cluster.EXPECT().Resources(gomock.Any()).Return(clusterResourcesFixture(), nil)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"get"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"web", "db", "local"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("expected %q in output:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "node/pve1") {
		t.Errorf("node resources should be excluded:\n%s", out.String())
	}
}

func TestResourcesCommandTypeAndStatusFilterJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupResourcesMocks(t, ctrl)
	cluster := mocks.NewMockClusterInterface(ctrl)
	client.EXPECT().Cluster(gomock.Any()).Return(cluster, nil)
	cluster.EXPECT().Resources(gomock.Any(), "vm").Return(clusterResourcesFixture()[:2], nil)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"get", "--type", "vm", "--status", "running", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	var summaries []resourceSummary
	if err := json.Unmarshal(out.Bytes(), &summaries); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if len(summaries) != 1 || summaries[0].Name != "web" || summaries[0].Type != "vm" {
		t.Fatalf("unexpected summaries: %+v", summaries)
	}
}

func TestResourcesCommandRejectsUnknownType(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupResourcesMocks(t, ctrl)

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"get", "--type", "disk"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("expected unsupported-type error, got %v", err)
	}
}
