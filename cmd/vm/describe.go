package vm

import (
	"context"
	"fmt"
	"io"
	"log"
	"proxmox-cli/cmd/utility"
	"reflect"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a virtual machine",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		node, _ := cmd.Flags().GetString("node")
		vmID, _ := cmd.Flags().GetInt("id")
		if node == "" || vmID == 0 {
			fmt.Fprintln(out, "Node name and VM ID must be provided. Use --node and --id flags.")
			return
		}

		// Fetch and display the VM details
		err := describeVirtualMachine(out, node, vmID)
		if err != nil {
			fmt.Fprintf(out, "Error describing virtual machine: %v\n", err)
			return
		}
	},
}

func init() {
	describeCmd.Flags().StringP("node", "n", "", "Node name")
	describeCmd.Flags().IntP("id", "i", 0, "VM ID")
	err := describeCmd.MarkFlagRequired("node")
	if err != nil {
		log.Fatal(err)
	}
	err = describeCmd.MarkFlagRequired("id")
	if err != nil {
		log.Fatal(err)
	}
}

func describeVirtualMachine(out io.Writer, node string, vmID int) error {
	err := utility.CheckIfAuthPresent()
	if err != nil {
		return err
	}

	client := utility.GetClient()

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		return err
	}

	_, err = retrievedNode.VirtualMachine(context.Background(), vmID)
	if err != nil {
		return err
	}

	// TODO: Need to refactor this to work with VirtualMachineInterface
	// For now, just print a simple message
	fmt.Fprintf(out, "VM %d details would be shown here\n", vmID)

	return nil
}

func printVMAttributes(out io.Writer, vm *proxmox.VirtualMachine) {
	var vmConfigField reflect.StructField
	var vmConfigValue reflect.Value
	v := reflect.ValueOf(vm).Elem()
	t := v.Type()
	fmt.Fprintln(out, "VM Details:")
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
			fmt.Fprintf(out, "%s: %s\n", field.Name, formatUptime(value.Uint()))
		} else {
			fmt.Fprintf(out, "%s: %v\n", field.Name, value.Interface())
		}
	}
	// Print VirtualMachineConfig at the end
	if !vmConfigValue.IsNil() {
		fmt.Fprintf(out, "%s:\n", vmConfigField.Name)

		for i := 0; i < vmConfigValue.Elem().NumField(); i++ {
			field := vmConfigValue.Elem().Type().Field(i)
			fieldValue := vmConfigValue.Elem().Field(i)

			// Skip empty strings, maps, and slices
			if (field.Type.Kind() == reflect.String && fieldValue.String() == "") ||
				(field.Type.Kind() == reflect.Map && fieldValue.Len() == 0) ||
				(field.Type.Kind() == reflect.Slice && fieldValue.Len() == 0) {
				continue
			}

			fmt.Fprintf(out, "    %s: %+v\n", field.Name, fieldValue.Interface())
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
