package utility

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// DefaultContext is the context name legacy single-cluster configurations
// are migrated into.
const DefaultContext = "default"

var (
	activeContextOverride string
	contextNamePattern    = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// ValidateContextName enforces names that are safe to embed in viper key
// paths (viper lowercases keys and splits on dots).
func ValidateContextName(name string) error {
	if !contextNamePattern.MatchString(name) {
		return fmt.Errorf("invalid context name %q; use lowercase letters, digits, and dashes", name)
	}
	return nil
}

// SetActiveContextOverride selects the context for this invocation, as set
// by the root --context flag. An empty name clears the override.
func SetActiveContextOverride(name string) {
	activeContextOverride = name
}

// ActiveContext returns the context this invocation operates on: the
// --context override if given, then the persisted current_context, then the
// default.
func ActiveContext() string {
	if activeContextOverride != "" {
		return activeContextOverride
	}
	if current := viper.GetString("current_context"); current != "" {
		return current
	}
	return DefaultContext
}

func contextKey(key string) string {
	return "contexts." + ActiveContext() + "." + key
}

// ContextString returns a config value for the active context. The default
// context falls back to the legacy flat layout so pre-context configs keep
// working until the next write migrates them.
func ContextString(key string) string {
	if value := viper.GetString(contextKey(key)); value != "" {
		return value
	}
	if ActiveContext() == DefaultContext {
		return viper.GetString(key)
	}
	return ""
}

// ContextBool is ContextString's boolean counterpart.
func ContextBool(key string) bool {
	if viper.IsSet(contextKey(key)) {
		return viper.GetBool(contextKey(key))
	}
	if ActiveContext() == DefaultContext {
		return viper.GetBool(key)
	}
	return false
}

// SetContextValue stores a config value under the active context.
func SetContextValue(key string, value any) {
	viper.Set(contextKey(key), value)
}

// ContextInfo describes one configured context.
type ContextInfo struct {
	Name      string `json:"name"`
	ServerURL string `json:"server_url"`
	Current   bool   `json:"current"`
}

// ListContexts enumerates configured contexts, including a legacy flat
// configuration as the default context.
func ListContexts() []ContextInfo {
	names := map[string]struct{}{}
	for _, key := range viper.AllKeys() {
		rest, found := strings.CutPrefix(key, "contexts.")
		if !found {
			continue
		}
		name, _, _ := strings.Cut(rest, ".")
		if name != "" {
			names[name] = struct{}{}
		}
	}
	if viper.GetString("server_url") != "" {
		names[DefaultContext] = struct{}{}
	}

	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	active := ActiveContext()
	infos := make([]ContextInfo, 0, len(sorted))
	for _, name := range sorted {
		serverURL := viper.GetString("contexts." + name + ".server_url")
		if serverURL == "" && name == DefaultContext {
			serverURL = viper.GetString("server_url")
		}
		if serverURL == "" {
			continue
		}
		infos = append(infos, ContextInfo{Name: name, ServerURL: serverURL, Current: name == active})
	}
	return infos
}

// ContextExists reports whether a context with a configured server exists.
func ContextExists(name string) bool {
	for _, info := range ListContexts() {
		if info.Name == name {
			return true
		}
	}
	return false
}

// DeleteContext blanks every value of the named context; WriteConfig prunes
// the emptied context from the persisted file.
func DeleteContext(name string) {
	viper.Set("contexts."+name, map[string]any{})
	for _, key := range viper.AllKeys() {
		if strings.HasPrefix(key, "contexts."+name+".") {
			viper.Set(key, "")
		}
	}
	if name == DefaultContext {
		// A legacy flat configuration is the default context.
		for _, key := range []string{"server_url", "ca_cert", "insecure"} {
			if viper.IsSet(key) {
				viper.Set(key, "")
			}
		}
		clearCredentialSection("auth_ticket")
		clearCredentialSection("api_token")
	}
}
