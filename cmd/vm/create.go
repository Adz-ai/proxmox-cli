package vm

import (
	"context"
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

func newCreateVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new virtual machine from a YAML spec file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			specFile, err := cmd.Flags().GetString("spec")
			if err != nil {
				return fmt.Errorf("get spec flag: %w", err)
			}
			node, err := cmd.Flags().GetString("node")
			if err != nil {
				return fmt.Errorf("get node flag: %w", err)
			}
			id, err := cmd.Flags().GetInt("id")
			if err != nil {
				return fmt.Errorf("get id flag: %w", err)
			}
			node = strings.TrimSpace(node)
			specFile = strings.TrimSpace(specFile)
			if node == "" {
				return fmt.Errorf("validate node: node cannot be empty")
			}
			if specFile == "" {
				return fmt.Errorf("validate spec: spec path cannot be empty")
			}
			if id < 0 {
				return fmt.Errorf("validate id: id must be positive")
			}

			spec, err := readYAMLSpec(specFile)
			if err != nil {
				return fmt.Errorf("read VM spec %q: %w", specFile, err)
			}

			vmOptions, err := mapToVMOptions(spec)
			if err != nil {
				return fmt.Errorf("map VM spec to options: %w", err)
			}

			createdID, err := createVirtualMachine(cmd.Context(), node, id, vmOptions, utility.TaskTimeout(cmd))
			if err != nil {
				return fmt.Errorf("create virtual machine on node %q: %w", node, err)
			}

			fmt.Fprintf(out, "Virtual machine %d created successfully.\n", createdID)
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node to create the virtual machine")
	cmd.Flags().StringP("spec", "s", "", "Path to the YAML spec file")
	cmd.Flags().IntP("id", "i", 0, "ID of the virtual machine (omit to auto-assign the next free ID)")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("spec"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")

	return cmd
}

func readYAMLSpec(filename string) (map[string]interface{}, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(file, &spec); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	return spec, nil
}

func mapToVMOptions(spec map[string]interface{}) ([]proxmox.VirtualMachineOption, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("spec cannot be empty")
	}

	keys := make([]string, 0, len(spec))
	for key := range spec {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			return nil, fmt.Errorf("spec option name cannot be empty")
		}
		if strings.EqualFold(normalizedKey, "vmid") || strings.EqualFold(normalizedKey, "node") {
			return nil, fmt.Errorf("spec cannot override %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	options := make([]proxmox.VirtualMachineOption, 0, len(keys))
	for _, key := range keys {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  key,
			Value: spec[key],
		})
	}

	return options, nil
}

func createVirtualMachine(ctx context.Context, node string, vmID int, options []proxmox.VirtualMachineOption, timeout time.Duration) (int, error) {
	client, err := utility.AuthenticatedClient()
	if err != nil {
		return 0, fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	vmID, err = utility.ResolveVMID(ctx, client, vmID)
	if err != nil {
		return 0, err
	}

	retrievedNode, err := client.Node(ctx, node)
	if err != nil {
		return 0, fmt.Errorf("get node %q: %w", node, err)
	}

	task, err := retrievedNode.NewVirtualMachine(ctx, vmID, options...)
	if err != nil {
		return 0, fmt.Errorf("start create task: %w", err)
	}
	if err := utility.WaitForTask(ctx, task, timeout); err != nil {
		return 0, fmt.Errorf("wait for create task: %w", err)
	}

	return vmID, nil
}
