package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
)

func setupContextConfig(t *testing.T) string {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Cleanup(func() { utility.SetActiveContextOverride("") })
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("PROXMOX_CLI_CONFIG", path)
	return path
}

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestLegacyConfigMigratesOnWrite(t *testing.T) {
	path := setupContextConfig(t)
	if err := os.WriteFile(path, []byte(`{
  "server_url": "https://legacy.example.com:8006",
  "auth_ticket": {
    "ticket": "PVE:root@pam:legacy",
    "CSRFPreventionToken": "legacy-token"
  }
}
`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Any write migrates the layout; logout is the simplest one.
	if _, err := runRoot(t, "auth", "logout"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `"contexts"`) || !strings.Contains(content, `"current_context": "default"`) {
		t.Fatalf("expected migrated layout:\n%s", content)
	}
	if !strings.Contains(content, `"server_url": "https://legacy.example.com:8006"`) {
		t.Fatalf("server URL lost in migration:\n%s", content)
	}
	fresh := viper.New()
	fresh.SetConfigFile(path)
	fresh.SetConfigType("json")
	if err := fresh.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if fresh.GetString("server_url") != "" {
		t.Fatalf("legacy top-level server_url should be removed:\n%s", content)
	}
	if fresh.GetString("contexts.default.server_url") != "https://legacy.example.com:8006" {
		t.Fatalf("migrated context missing server URL:\n%s", content)
	}
}

func TestContextFlagIsolatesClusters(t *testing.T) {
	setupContextConfig(t)
	viper.Set("contexts.homelab.server_url", "https://homelab.example.com:8006")
	viper.Set("contexts.work.server_url", "https://work.example.com:8006")
	viper.Set("current_context", "homelab")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	out, err := runRoot(t, "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Context: homelab") || !strings.Contains(out, "https://homelab.example.com:8006") {
		t.Fatalf("expected current context homelab:\n%s", out)
	}

	out, err = runRoot(t, "--context", "work", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Context: work") || !strings.Contains(out, "https://work.example.com:8006") {
		t.Fatalf("expected work context via flag:\n%s", out)
	}
}

func TestContextUseSwitchesAndPersists(t *testing.T) {
	path := setupContextConfig(t)
	viper.Set("contexts.homelab.server_url", "https://homelab.example.com:8006")
	viper.Set("contexts.work.server_url", "https://work.example.com:8006")
	viper.Set("current_context", "homelab")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	out, err := runRoot(t, "context", "use", "work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `Switched to context "work"`) {
		t.Fatalf("unexpected output:\n%s", out)
	}

	fresh := viper.New()
	fresh.SetConfigFile(path)
	fresh.SetConfigType("json")
	if err := fresh.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if fresh.GetString("current_context") != "work" {
		t.Fatalf("current_context not persisted")
	}

	viper.Reset()
	out, err = runRoot(t, "context", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "*   work") {
		t.Fatalf("expected work marked current:\n%s", out)
	}

	viper.Reset()
	if _, err := runRoot(t, "context", "use", "ghost"); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected unknown-context error, got %v", err)
	}
}

func TestContextDelete(t *testing.T) {
	path := setupContextConfig(t)
	viper.Set("contexts.homelab.server_url", "https://homelab.example.com:8006")
	viper.Set("contexts.work.server_url", "https://work.example.com:8006")
	viper.Set("contexts.work.auth_ticket.ticket", "work-secret")
	viper.Set("current_context", "homelab")
	if err := utility.WriteConfig(); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	if _, err := runRoot(t, "context", "delete", "homelab"); err == nil || !strings.Contains(err.Error(), "active context") {
		t.Fatalf("expected active-context protection, got %v", err)
	}

	viper.Reset()
	if _, err := runRoot(t, "context", "delete", "work"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "work") {
		t.Fatalf("deleted context still in file:\n%s", data)
	}
	if !strings.Contains(string(data), "homelab") {
		t.Fatalf("surviving context lost:\n%s", data)
	}
}

func TestContextFlagRejectsInvalidName(t *testing.T) {
	setupContextConfig(t)
	_, err := runRoot(t, "--context", "Bad.Name", "status")
	if err == nil || !strings.Contains(err.Error(), "invalid context name") {
		t.Fatalf("expected invalid-name error, got %v", err)
	}
}
