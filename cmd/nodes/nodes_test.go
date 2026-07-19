package nodes

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
)

func setupAuthenticatedMocks(t *testing.T, ctrl *gomock.Controller) *mocks.MockProxmoxClientInterface {
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

func TestDescribeUnknownNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "ghost").Return(node, nil)
	client.EXPECT().Nodes(ctx).Return(proxmox.NodeStatuses{
		&proxmox.NodeStatus{Node: "pve", Status: "online"},
	}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"describe", "-n", "ghost"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found in cluster") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestGetJSONOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)

	client.EXPECT().Nodes(gomock.Any()).Return(proxmox.NodeStatuses{
		&proxmox.NodeStatus{Node: "pve1", Status: "online", Type: "node", Uptime: 86400},
		&proxmox.NodeStatus{Node: "pve2", Status: "offline", Type: "node"},
	}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var summaries []nodeSummary
	if err := json.Unmarshal(out.Bytes(), &summaries); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if len(summaries) != 2 || summaries[0].Node != "pve1" || summaries[1].Status != "offline" {
		t.Fatalf("unexpected summaries: %+v", summaries)
	}
}

func TestGetRejectsUnknownOutputFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupAuthenticatedMocks(t, ctrl)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "-o", "xml"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("expected unsupported-format error, got %v", err)
	}
}
