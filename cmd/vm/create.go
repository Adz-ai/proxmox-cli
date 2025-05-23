package vm

import (
	"context"
	"fmt"
	"log"
	"os"
	"proxmox-cli/cmd/utility"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var createVMCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new virtual machine from a YAML spec file",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		
		specFile, _ := cmd.Flags().GetString("spec")
		node, _ := cmd.Flags().GetString("node")
		id, _ := cmd.Flags().GetInt("id")

		// Read and parse the YAML spec file
		spec, err := readYAMLSpec(specFile)
		if err != nil {
			fmt.Fprintf(out, "Error reading spec file: %v\n", err)
			return
		}

		// Map the values to a slice of VirtualMachineOption
		vmOptions, err := mapToVMOptions(spec)
		if err != nil {
			fmt.Fprintf(out, "Error mapping spec to VM options: %v\n", err)
			return
		}

		// Create the VM
		err = createVirtualMachine(node, id, vmOptions)
		if err != nil {
			fmt.Fprintf(out, "Error creating virtual machine: %v\n", err)
			return
		}

		fmt.Fprintln(out, "Virtual machine created successfully.")
	},
}

func init() {
	createVMCmd.Flags().StringP("node", "n", "", "Node to create the virtual machine")
	createVMCmd.Flags().StringP("spec", "s", "", "Path to the YAML spec file")
	createVMCmd.Flags().IntP("id", "i", 1, "ID of the virtual machine to be created")
	err := createVMCmd.MarkFlagRequired("node")
	if err != nil {
		log.Fatal(err)
	}
	err = createVMCmd.MarkFlagRequired("spec")
	if err != nil {
		log.Fatal(err)
	}
	err = createVMCmd.MarkFlagRequired("id")
	if err != nil {
		log.Fatal(err)
	}
}

func readYAMLSpec(filename string) (map[string]interface{}, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	err = yaml.Unmarshal(file, &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func mapToVMOptions(spec map[string]interface{}) ([]proxmox.VirtualMachineOption, error) {
	var options []proxmox.VirtualMachineOption

	for key, value := range spec {
		options = append(options, proxmox.VirtualMachineOption{
			Name:  key,
			Value: value,
		})
	}

	return options, nil
}

func createVirtualMachine(node string, vmId int, options []proxmox.VirtualMachineOption) error {
	err := utility.CheckIfAuthPresent()
	if err != nil {
		return err
	}

	client := utility.GetClient()

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		return err
	}

	// Create the virtual machine
	task, err := retrievedNode.NewVirtualMachine(context.Background(), vmId, options...)
	if err != nil {
		return err
	}

	err = task.WaitFor(context.Background(), 10)
	if err != nil {
		return err
	}

	if !task.IsSuccessful {
		return fmt.Errorf("VM create failed")
	}

	return nil
}
