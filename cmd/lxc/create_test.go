package lxc

import (
	"testing"
)

func TestContainerOptionsFromSpec(t *testing.T) {
	if _, err := containerOptionsFromSpec(map[string]any{}); err == nil {
		t.Error("expected error when ostemplate is missing")
	}
	if _, err := containerOptionsFromSpec(map[string]any{"ostemplate": "  "}); err == nil {
		t.Error("expected error when ostemplate is blank")
	}
	if _, err := containerOptionsFromSpec(map[string]any{"ostemplate": "local:vztmpl/debian.tar.zst", "vmid": 200}); err == nil {
		t.Error("expected error when spec overrides vmid")
	}
	if _, err := containerOptionsFromSpec(map[string]any{"ostemplate": "local:vztmpl/debian.tar.zst", "bogus": 1}); err == nil {
		t.Error("expected error for unsupported key")
	}

	options, err := containerOptionsFromSpec(map[string]any{
		"ostemplate": "local:vztmpl/debian.tar.zst",
		"net0":       "name=eth0,bridge=vmbr0",
		"cores":      2,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(options))
	for _, option := range options {
		got = append(got, option.Name)
	}
	want := []string{"cores", "net0", "ostemplate"}
	if len(got) != len(want) {
		t.Fatalf("option names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("option names = %v, want %v (sorted)", got, want)
		}
	}
}

func TestIndexedContainerCreateKey(t *testing.T) {
	for _, key := range []string{"net0", "mp12", "dev3"} {
		if !indexedContainerCreateKey(key) {
			t.Errorf("indexedContainerCreateKey(%q) = false, want true", key)
		}
	}
	for _, key := range []string{"net", "netX", "mp1a", "memory", "0net"} {
		if indexedContainerCreateKey(key) {
			t.Errorf("indexedContainerCreateKey(%q) = true, want false", key)
		}
	}
}

func TestValidateContainerTarget(t *testing.T) {
	if err := validateContainerTarget("pve", 100); err != nil {
		t.Errorf("valid target rejected: %v", err)
	}
	if err := validateContainerTarget("  ", 100); err == nil {
		t.Error("expected error for blank node")
	}
	if err := validateContainerTarget("pve", 0); err == nil {
		t.Error("expected error for nonpositive vmid")
	}
}
