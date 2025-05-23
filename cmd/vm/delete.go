package vm

import (
	"context"
	"fmt"
	"log"
	"proxmox-cli/cmd/utility"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete virtual machine",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		node, _ := cmd.Flags().GetString("node")
		id, _ := cmd.Flags().GetInt("id")
		
		err := deleteVm(node, id)
		if err != nil {
			fmt.Fprintf(out, "Error deleting VM: %v\n", err)
			return
		}
		
		fmt.Fprintf(out, "VM %d deleted successfully\n", id)
	},
}

func init() {
	deleteCmd.Flags().StringP("node", "n", "", "node to delete VM from (required)")
	deleteCmd.Flags().IntP("id", "i", 0, "id for VM to delete (required)")
	err := deleteCmd.MarkFlagRequired("node")
	if err != nil {
		log.Fatal(err)
	}
	err = deleteCmd.MarkFlagRequired("id")
	if err != nil {
		log.Fatal(err)
	}
}

func deleteVm(node string, id int) error {
	err := utility.CheckIfAuthPresent()
	if err != nil {
		return err
	}

	client := utility.GetClient()

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		return err
	}

	vmToDelete, err := retrievedNode.VirtualMachine(context.Background(), id)
	if err != nil {
		return err
	}

	task, err := vmToDelete.Delete(context.Background())
	if err != nil {
		return err
	}

	err = task.WaitFor(context.Background(), 5)
	if err != nil {
		return err
	}

	if !task.IsSuccessful {
		return fmt.Errorf("VM delete failed")
	}

	return nil
}
