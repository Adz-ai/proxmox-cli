package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/Adz-ai/proxmox-cli/test/mocks"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"
)

func TestNewRootCmdResetsFlagState(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("PROXMOX_CLI_CONFIG", filepath.Join(t.TempDir(), "config.json"))

	first := NewRootCmd()
	first.SetArgs([]string{"status", "--verbose"})
	if err := first.Execute(); err != nil {
		t.Fatal(err)
	}

	second := NewRootCmd()
	status, _, err := second.Find([]string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	verbose, err := status.Flags().GetBool("verbose")
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		t.Fatal("verbose flag leaked between command roots")
	}
}

func TestNewRootCmdBuildsFreshCommandTrees(t *testing.T) {
	first := NewRootCmd()
	second := NewRootCmd()
	firstGet, _, err := first.Find([]string{"lxc", "get"})
	if err != nil {
		t.Fatal(err)
	}
	secondGet, _, err := second.Find([]string{"lxc", "get"})
	if err != nil {
		t.Fatal(err)
	}
	if firstGet == secondGet {
		t.Fatal("command trees share leaf command pointers")
	}
}

func TestStatusDoesNotCreateConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	dir := filepath.Join(t.TempDir(), "config")
	t.Setenv("PROXMOX_CLI_CONFIG", filepath.Join(dir, "config.json"))

	root := NewRootCmd()
	root.SetArgs([]string{"status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("status created config directory: %v", err)
	}
}

func TestInitPreservesTLSSettingsUnlessChanged(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("PROXMOX_CLI_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	viper.Set("server_url", "https://old.example.com:8006")
	viper.Set("ca_cert", "/private/ca.pem")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd()
	root.SetIn(strings.NewReader("https://new.example.com:8006\n"))
	root.SetArgs([]string{"init", "--force"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := viper.GetString("ca_cert"); got != "/private/ca.pem" {
		t.Fatalf("ca_cert = %q, want preserved value", got)
	}
}

func TestCommandContextIsPropagated(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("PROXMOX_CLI_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	viper.Set("server_url", "https://pve.example.com:8006")
	viper.Set("auth_ticket.ticket", "ticket")
	viper.Set("auth_ticket.CSRFPreventionToken", "token")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	client := mocks.NewMockProxmoxClientInterface(ctrl)
	client.EXPECT().Version(gomock.Any()).DoAndReturn(func(ctx context.Context) (*proxmox.Version, error) {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("command context was not canceled: %v", ctx.Err())
		}
		return nil, ctx.Err()
	})
	utility.SetClientFactory(func() interfaces.ProxmoxClientInterface { return client })
	t.Cleanup(utility.ResetClientFactory)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := NewRootCmd()
	root.SetArgs([]string{"status", "--verbose"})
	if err := root.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteContext() error = %v, want context canceled", err)
	}
}
