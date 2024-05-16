package vm

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"proxmox-cli/cmd/utility"
)

var createVMCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new virtual machine from a YAML spec file",
	Run: func(cmd *cobra.Command, args []string) {
		specFile, _ := cmd.Flags().GetString("spec")
		node, _ := cmd.Flags().GetString("node")
		id, _ := cmd.Flags().GetInt("id")

		// Read and parse the YAML spec file
		spec, err := readYAMLSpec(specFile)
		if err != nil {
			log.Fatalf("Error reading spec file: %v", err)
		}

		// Map the values to a slice of VirtualMachineOption
		vmOptions, err := mapToVMOptions(spec)
		if err != nil {
			log.Fatalf("Error mapping spec to VM options: %v", err)
		}

		// Create the VM
		err = createVirtualMachine(node, id, vmOptions)
		if err != nil {
			log.Fatalf("Error creating virtual machine: %v", err)
		}

		fmt.Println("Virtual machine created successfully.")
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
	Cmd.AddCommand(createVMCmd)
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
	utility.CheckIfAuthPresent()

	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", viper.GetString("server_url")),
		proxmox.WithSession(viper.Sub("auth_ticket").GetString("ticket"), viper.Sub("auth_ticket").GetString("CSRFPreventionToken")),
	)

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal("VM Create failed")
	}

	return nil
}
