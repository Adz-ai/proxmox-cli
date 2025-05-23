package nodes

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
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
	err := utility.CheckIfAuthPresent()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := utility.GetClient()

	nodes, err := client.Nodes(context.Background())
	if err != nil {
		log.Fatalf("Error fetching nodes: %v", err)
	}
	
	fmt.Println("Nodes in cluster:")
	fmt.Println("=================")
	fmt.Printf("%-15s %-10s %-8s %-12s\n", "Node", "Status", "Type", "Uptime")
	fmt.Printf("%-15s %-10s %-8s %-12s\n", "----", "------", "----", "------")
	for _, node := range nodes {
		uptime := "N/A"
		if node.Uptime > 0 {
			days := node.Uptime / 86400
			hours := (node.Uptime % 86400) / 3600
			uptime = fmt.Sprintf("%dd %dh", days, hours)
		}
		fmt.Printf("%-15s %-10s %-8s %-12s\n", 
			node.Node,
			node.Status,
			node.Type,
			uptime)
	}
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
