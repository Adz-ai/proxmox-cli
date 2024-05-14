package nodes

import (
	"context"
	"fmt"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"proxmox-cli/cmd/utility"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		viewNodesDetails()
	},
}

func init() {
	Cmd.AddCommand(getCmd)
}

func viewNodesDetails() {
	utility.CheckIfAuthPresent()

	client := proxmox.NewClient(fmt.Sprintf("%s/api2/json", viper.GetString("server_url")),
		proxmox.WithSession(viper.Sub("auth_ticket").GetString("ticket"), viper.Sub("auth_ticket").GetString("CSRFPreventionToken")),
	)

	nodes, err := client.Nodes(context.Background())
	if err != nil {
		log.Fatalf("Error fetching nodes: %v", err)
	}
	for _, node := range nodes {
		fmt.Printf("Node: %s\n", node.Node)
	}
}
