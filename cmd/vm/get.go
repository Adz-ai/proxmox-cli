package vm

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"proxmox-cli/cmd/utility"
	"time"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get virtual machines",
	Run: func(cmd *cobra.Command, args []string) {
		viewVMs()
	},
}

func init() {
	Cmd.AddCommand(getCmd)
}

func viewVMs() {

	utility.CheckIfAuthPresent()

	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", viper.GetString("server_url")),
		proxmox.WithSession(viper.Sub("auth_ticket").GetString("ticket"), viper.Sub("auth_ticket").GetString("CSRFPreventionToken")),
	)

	nodes, err := client.Nodes(context.Background())
	if err != nil {
		log.Fatalf("Error fetching nodes: %v", err)
	}

	for _, node := range nodes {
		fmt.Println("Node: " + node.Node)
		n, err := client.Node(context.Background(), node.Node)
		if err != nil {
			log.Fatalf("Error fetching node %s: %v", node.Node, err)
		}
		vms, err := n.VirtualMachines(context.Background())

		if err != nil {
			log.Fatalf("Error fetching VMs for node %s: %v", node.Node, err)
		}
		for _, vm := range vms {
			uptime := time.Duration(vm.Uptime) * time.Second
			days := int(uptime.Hours() / 24)
			hours := int(uptime.Hours()) % 24
			minutes := int(uptime.Minutes()) % 60
			seconds := int(uptime.Seconds()) % 60
			fmt.Printf("VM: %s\n ID: %d\n Status: %s\n Uptime: %d days, %d hours, %d minutes, %d seconds\n", vm.Name, vm.VMID, vm.Status, days, hours, minutes, seconds)
		}
	}

}
