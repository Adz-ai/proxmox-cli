/*
Copyright © 2024 Adarssh Athithan
*/
package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"proxmox-cli/cmd/auth"
	"proxmox-cli/cmd/lxc"
	"proxmox-cli/cmd/nodes"
	"proxmox-cli/cmd/vm"
	"runtime"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proxmox-cli",
	Short: "Command-line interface for Proxmox VE",
	Long: `🚀 Proxmox CLI - Manage your Proxmox Virtual Environment from the terminal

A powerful command-line tool for managing Proxmox VE resources including:
• Virtual Machines (VMs)
• LXC Containers  
• Cluster Nodes
• Storage and Networks

Get started:
  proxmox-cli init                    # Configure connection
  proxmox-cli auth login -u root@pam  # Authenticate
  proxmox-cli status                  # Check connection`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(nodes.Cmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(vm.VMCmd)
	rootCmd.AddCommand(lxc.LXCCmd)
}

func initConfig() {
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("HOMEPATH")
	} else {
		homeDir = os.Getenv("HOME")
	}
	configPath := filepath.Join(homeDir, ".proxmox-cli")
	configName := "config"
	configType := "json"

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// Attempt to read the config, if it doesn't exist, create empty config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// The Config file does not exist; create empty one
			err := os.MkdirAll(configPath, os.ModePerm)
			if err != nil {
				return
			}
			// Create empty config file without prompting
			err = viper.SafeWriteConfig()
			if err != nil {
				return
			}
		}
	}
}

