/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package nodes

import (
	"github.com/spf13/cobra"
)

// NodesCmd represents the nodes command
var NodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Nodes is a palette that contains node based commands",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
