package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestPromptSecretFromReader(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetIn(strings.NewReader("  s3cret \n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	secret, err := promptSecret(cmd, "Secret: ")
	if err != nil {
		t.Fatal(err)
	}
	if secret != "s3cret" {
		t.Fatalf("promptSecret = %q, want %q", secret, "s3cret")
	}
}

func TestTokenRejectsMalformedTokenID(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader("secret\n"))
	cmd.SetArgs([]string{"token", "-t", "missing-separator"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "user@realm!tokenname") {
		t.Fatalf("expected malformed token ID error, got %v", err)
	}
}

func TestTokenRequiresConfiguredServer(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader("secret\n"))
	cmd.SetArgs([]string{"token", "-t", "root@pam!ci"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "server URL is not configured") {
		t.Fatalf("expected unconfigured-server error, got %v", err)
	}
}

func TestLogoutWhenNotLoggedIn(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"logout"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Not currently logged in") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}
