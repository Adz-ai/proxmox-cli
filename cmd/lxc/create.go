package lxc

import (
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new LXC container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			client, err := utility.AuthenticatedClient()
			if err != nil {
				return fmt.Errorf("authenticate Proxmox client: %w", err)
			}

			nodeName, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("read node flag: %w", err)
			}
			vmid, err := cmd.Flags().GetInt("vmid")
			if err != nil {
				return fmt.Errorf("read vmid flag: %w", err)
			}
			specFile, err := cmd.Flags().GetString("spec")
			if err != nil {
				return fmt.Errorf("read spec flag: %w", err)
			}
			if err := validateContainerTarget(nodeName, vmid); err != nil {
				return err
			}
			if strings.TrimSpace(specFile) == "" {
				return fmt.Errorf("spec path cannot be empty")
			}

			data, err := os.ReadFile(specFile)
			if err != nil {
				return fmt.Errorf("read spec file %q: %w", specFile, err)
			}

			var spec map[string]any
			if err := yaml.Unmarshal(data, &spec); err != nil {
				return fmt.Errorf("parse spec file %q: %w", specFile, err)
			}
			options, err := containerOptionsFromSpec(spec)
			if err != nil {
				return fmt.Errorf("validate container spec: %w", err)
			}

			node, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}

			task, err := node.NewContainer(ctx, vmid, options...)
			if err != nil {
				return fmt.Errorf("create container %d: %w", vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, 10*time.Minute); err != nil {
				return fmt.Errorf("create container %d: %w", vmid, err)
			}

			fmt.Fprintln(out, "Container created successfully")
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node name")
	cmd.Flags().IntP("vmid", "i", 0, "Container ID")
	cmd.Flags().StringP("spec", "s", "", "YAML specification file")
	for _, flag := range []string{"node", "vmid", "spec"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	return cmd
}

var containerCreateKeys = map[string]struct{}{
	"arch": {}, "bwlimit": {}, "cmode": {}, "console": {}, "cores": {},
	"cpulimit": {}, "cpuunits": {}, "debug": {}, "description": {}, "features": {},
	"force": {}, "ha-managed": {}, "hookscript": {}, "hostname": {}, "ignore-unpack-errors": {},
	"lock": {}, "memory": {}, "nameserver": {}, "onboot": {}, "ostemplate": {},
	"ostype": {}, "password": {}, "pool": {}, "protection": {}, "restore": {},
	"rootfs": {}, "searchdomain": {}, "ssh-public-keys": {}, "start": {},
	"startup": {}, "storage": {}, "swap": {}, "tags": {}, "template": {},
	"timezone": {}, "tty": {}, "unique": {}, "unprivileged": {},
}

func containerOptionsFromSpec(spec map[string]any) ([]proxmox.ContainerOption, error) {
	template, ok := spec["ostemplate"]
	if !ok || template == nil {
		return nil, fmt.Errorf("ostemplate is required")
	}
	if value, ok := template.(string); !ok || strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("ostemplate must be a nonempty string")
	}

	keys := make([]string, 0, len(spec))
	for key := range spec {
		if key == "vmid" || key == "node" {
			return nil, fmt.Errorf("%q must be provided as a command flag, not in the spec", key)
		}
		if _, ok := containerCreateKeys[key]; !ok && !indexedContainerCreateKey(key) {
			return nil, fmt.Errorf("unsupported container create key %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	options := make([]proxmox.ContainerOption, 0, len(keys))
	for _, key := range keys {
		options = append(options, proxmox.ContainerOption{Name: key, Value: spec[key]})
	}
	return options, nil
}

func indexedContainerCreateKey(key string) bool {
	for _, prefix := range []string{"dev", "mp", "net"} {
		if suffix := strings.TrimPrefix(key, prefix); suffix != key && suffix != "" {
			for _, digit := range suffix {
				if digit < '0' || digit > '9' {
					return false
				}
			}
			return true
		}
	}
	return false
}
