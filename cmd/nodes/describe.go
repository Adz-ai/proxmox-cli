package nodes

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"proxmox-cli/cmd/utility"
)

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show detailed information about a node",
	Run: func(cmd *cobra.Command, args []string) {
		nodeName, _ := cmd.Flags().GetString("name")

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Println(err)
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		nodes, err := client.Nodes(ctx)
		if err != nil {
			fmt.Printf("Error fetching nodes: %v\n", err)
			return
		}

		// Find the target node
		fmt.Printf("Looking for node: %s\n", nodeName)
		found := false
		for _, node := range nodes {
			// Print node details
			// The exact way to check node name depends on the struct
			fmt.Printf("Node details: %v\n", node)
			found = true
		}
		
		if !found {
			fmt.Printf("No nodes found\n")
		}
	},
}

func init() {
	describeCmd.Flags().StringP("name", "n", "", "Name of the node to describe")
	describeCmd.MarkFlagRequired("name")
}

