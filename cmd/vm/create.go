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
			if id <= 0 {
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

			if err := createVirtualMachine(cmd.Context(), node, id, vmOptions); err != nil {
				return fmt.Errorf("create virtual machine %d on node %q: %w", id, node, err)
			}

			fmt.Fprintln(out, "Virtual machine created successfully.")
			return nil
		},
	}

	cmd.Flags().StringP("node", "n", "", "Node to create the virtual machine")
	cmd.Flags().StringP("spec", "s", "", "Path to the YAML spec file")
	cmd.Flags().IntP("id", "i", 0, "ID of the virtual machine to be created")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("spec"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}

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

func createVirtualMachine(ctx context.Context, node string, vmID int, options []proxmox.VirtualMachineOption) error {
	client, err := utility.AuthenticatedClient()
	if err != nil {
		return fmt.Errorf("authenticate Proxmox client: %w", err)
	}

	retrievedNode, err := client.Node(ctx, node)
	if err != nil {
		return fmt.Errorf("get node %q: %w", node, err)
	}

	task, err := retrievedNode.NewVirtualMachine(ctx, vmID, options...)
	if err != nil {
		return fmt.Errorf("start create task: %w", err)
	}
	if err := utility.WaitForTask(ctx, task, 10*time.Minute); err != nil {
		return fmt.Errorf("wait for create task: %w", err)
	}

	return nil
}
