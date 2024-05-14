/*
Copyright © 2024 Adarssh Athithan
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// NodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Manage Proxmox Nodes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(nodesCmd)
}
