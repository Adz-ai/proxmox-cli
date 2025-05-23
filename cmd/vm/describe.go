package vm

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"proxmox-cli/cmd/utility"
	"reflect"
	"time"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a virtual machine",
	Run: func(cmd *cobra.Command, args []string) {
		node, _ := cmd.Flags().GetString("node")
		vmID, _ := cmd.Flags().GetInt("id")
		if node == "" || vmID == 0 {
			log.Fatalf("Node name and VM ID must be provided. Use --node and --vmid flags.")
		}

		// Fetch and display the VM details
		err := describeVirtualMachine(node, vmID)
		if err != nil {
			log.Fatalf("Error describing virtual machine: %v", err)
		}
	},
}

func init() {
	describeCmd.Flags().StringP("node", "n", "", "Node name")
	describeCmd.Flags().IntP("id", "i", 0, "VM ID")
	describeCmd.MarkFlagRequired("node")
	describeCmd.MarkFlagRequired("id")
	VMCmd.AddCommand(describeCmd)
}

func describeVirtualMachine(node string, vmID int) error {
	utility.CheckIfAuthPresent()

	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", viper.GetString("server_url")),
		proxmox.WithSession(viper.Sub("auth_ticket").GetString("ticket"), viper.Sub("auth_ticket").GetString("CSRFPreventionToken")),
	)

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		log.Fatal(err)
	}

	vm, err := retrievedNode.VirtualMachine(context.Background(), vmID)
	if err != nil {
		return err
	}

	printVMAttributes(vm)

	return nil
}

func printVMAttributes(vm *proxmox.VirtualMachine) {
	var vmConfigField reflect.StructField
	var vmConfigValue reflect.Value
	v := reflect.ValueOf(vm).Elem()
	t := v.Type()
	fmt.Println("VM Details:")
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Name == "client" {
			continue
		}

		if field.Name == "VirtualMachineConfig" {
			vmConfigField = field
			vmConfigValue = value
			continue
		}

		if field.Name == "Uptime" {
			fmt.Printf("%s: %s\n", field.Name, formatUptime(value.Uint()))
		} else {
			fmt.Printf("%s: %v\n", field.Name, value.Interface())
		}
	}
	// Print VirtualMachineConfig at the end
	if !vmConfigValue.IsNil() {
		fmt.Printf("%s:\n", vmConfigField.Name)

		for i := 0; i < vmConfigValue.Elem().NumField(); i++ {
			field := vmConfigValue.Elem().Type().Field(i)
			fieldValue := vmConfigValue.Elem().Field(i)

			// Skip empty strings, maps, and slices
			if (field.Type.Kind() == reflect.String && fieldValue.String() == "") ||
				(field.Type.Kind() == reflect.Map && fieldValue.Len() == 0) ||
				(field.Type.Kind() == reflect.Slice && fieldValue.Len() == 0) {
				continue
			}

			fmt.Printf("    %s: %+v\n", field.Name, fieldValue.Interface())
		}
	}
}

func formatUptime(uptime uint64) string {
	duration := time.Duration(uptime) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours, minutes, seconds)
}
