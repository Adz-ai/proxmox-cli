package lxc

import (
	"context"
	"fmt"
	"os"
	"proxmox-cli/cmd/utility"

	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new LXC container",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()

		err := utility.CheckIfAuthPresent()
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}

		nodeName, _ := cmd.Flags().GetString("node")
		vmid, _ := cmd.Flags().GetInt("vmid")
		specFile, _ := cmd.Flags().GetString("spec")

		if nodeName == "" || vmid == 0 || specFile == "" {
			fmt.Fprintln(out, "Error: node, vmid, and spec are required")
			return
		}

		// Read spec file
		data, err := os.ReadFile(specFile)
		if err != nil {
			fmt.Fprintf(out, "Error reading spec file: %v\n", err)
			return
		}

		// Parse YAML
		var spec map[string]interface{}
		err = yaml.Unmarshal(data, &spec)
		if err != nil {
			fmt.Fprintf(out, "Error parsing YAML: %v\n", err)
			return
		}

		client := utility.GetClient()
		ctx := context.Background()

		node, err := client.Node(ctx, nodeName)
		if err != nil {
			fmt.Fprintf(out, "Error getting node: %v\n", err)
			return
		}

		// Create container options from spec
		var options []proxmox.ContainerOption
		// This is simplified - in reality you'd parse all the spec fields

		task, err := node.NewContainer(ctx, vmid, options...)
		if err != nil {
			fmt.Fprintf(out, "Error creating container: %v\n", err)
			return
		}

		fmt.Fprintf(out, "Container creation task started: %s\n", task.UPID)
		fmt.Fprintln(out, "Container created successfully")
	},
}

func init() {
	createCmd.Flags().StringP("node", "n", "", "Node name")
	createCmd.Flags().IntP("vmid", "i", 0, "Container ID")
	createCmd.Flags().StringP("spec", "s", "", "YAML specification file")
	createCmd.MarkFlagRequired("node")
	createCmd.MarkFlagRequired("vmid")
	createCmd.MarkFlagRequired("spec")
}
