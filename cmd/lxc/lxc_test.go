package lxc

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

func TestCloneCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Container(ctx, 200).Return(container, nil)
	container.EXPECT().Clone(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error) {
			if options.NewID != 201 || options.Hostname != "copy" {
				t.Errorf("unexpected clone options: %+v", options)
			}
			return 201, &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"clone", "-n", "pve", "-s", "200", "-t", "201", "--name", "copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Container 200 cloned to 201 successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestCloneRejectsSameSourceAndTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupAuthenticatedMocks(t, ctrl)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"clone", "-n", "pve", "-s", "200", "-t", "200"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "must differ from the source") {
		t.Fatalf("expected same-ID error, got %v", err)
	}
}

func TestSnapshotCreateRejectsBlankName(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupAuthenticatedMocks(t, ctrl)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"snapshot", "create", "-n", "pve", "-i", "200", "--name", "  "})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "snapshot name cannot be empty") {
		t.Fatalf("expected blank-name error, got %v", err)
	}
}

func TestShutdownForwardsForceAndGrace(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Container(ctx, 200).Return(container, nil)
	container.EXPECT().Shutdown(ctx, true, 30).Return(&proxmox.Task{IsSuccessful: true}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"shutdown", "-n", "pve", "-i", "200", "--force", "--grace-seconds", "30"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Container 200 shut down successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestCloneAutoAssignsTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)
	cluster := mocks.NewMockClusterInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Cluster(ctx).Return(cluster, nil)
	cluster.EXPECT().NextID(ctx).Return(201, nil)
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Container(ctx, 200).Return(container, nil)
	container.EXPECT().Clone(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, options *proxmox.ContainerCloneOptions) (int, *proxmox.Task, error) {
			if options.NewID != 201 {
				t.Errorf("clone target = %d, want auto-assigned 201", options.NewID)
			}
			return 201, &proxmox.Task{IsSuccessful: true}, nil
		})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"clone", "-n", "pve", "-s", "200"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Container 200 cloned to 201 successfully") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestSnapshotDeleteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Container(ctx, 200).Return(container, nil)
	container.EXPECT().DeleteSnapshot(ctx, "old").Return(&proxmox.Task{IsSuccessful: true}, nil)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"snapshot", "delete", "-n", "pve", "-i", "200", "--name", "old"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `Snapshot "old" of container 200 deleted successfully`) {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}

func TestDescribeJSONOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := setupAuthenticatedMocks(t, ctrl)
	node := mocks.NewMockNodeInterface(ctrl)
	container := mocks.NewMockContainerInterface(ctrl)

	ctx := gomock.Any()
	client.EXPECT().Node(ctx, "pve").Return(node, nil)
	node.EXPECT().Container(ctx, 200).Return(container, nil)
	container.EXPECT().Details().Return(interfaces.ContainerDetails{
		Name:   "web",
		Node:   "pve",
		Status: "running",
		CPUs:   2,
	})

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"describe", "-n", "pve", "-i", "200", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"vmid": 200`, `"name": "web"`, `"status": "running"`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("JSON output missing %s:\n%s", want, out.String())
		}
	}
}
