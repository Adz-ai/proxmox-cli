package vm

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"proxmox-cli/cmd/utility"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete virtual machine",
	Run: func(cmd *cobra.Command, args []string) {
		node, _ := cmd.Flags().GetString("node")
		id, _ := cmd.Flags().GetInt("id")
		deleteVm(node, id)
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
	VMCmd.AddCommand(deleteCmd)
}

func deleteVm(node string, id int) {
	utility.CheckIfAuthPresent()

	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", viper.GetString("server_url")),
		proxmox.WithSession(viper.Sub("auth_ticket").GetString("ticket"), viper.Sub("auth_ticket").GetString("CSRFPreventionToken")),
	)

	retrievedNode, err := client.Node(context.Background(), node)
	if err != nil {
		log.Fatal(err)
	}

	vmToDelete, err := retrievedNode.VirtualMachine(context.Background(), id)
	if err != nil {
		log.Fatal(err)
	}

	task, err := vmToDelete.Delete(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = task.WaitFor(context.Background(), 5)
	if err != nil {
		return
	}

	if !task.IsSuccessful {
		log.Fatal("VM delete failed")
	}

	fmt.Printf("VM %s deleted successfully\n", vmToDelete.Name)

}
