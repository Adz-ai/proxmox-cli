package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"proxmox-cli/cmd/auth"
	"proxmox-cli/cmd/lxc"
	"proxmox-cli/cmd/nodes"
	"proxmox-cli/cmd/vm"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// NewRootCmd creates a new root command for testing
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxmox-cli",
		Short: "Command-line interface for Proxmox VE",
		Long: rootCmd.Long,
	}
	
	// Add all subcommands
	cmd.AddCommand(initCmd)
	cmd.AddCommand(statusCmd)
	cmd.AddCommand(nodes.Cmd)
	cmd.AddCommand(auth.Cmd)
	cmd.AddCommand(vm.Cmd)
	cmd.AddCommand(lxc.Cmd)
	
	return cmd
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(nodes.Cmd)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(vm.Cmd)
	rootCmd.AddCommand(lxc.Cmd)
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

	// Attempt to read the config, if it doesn't exist, create an empty config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// The Config file does not exist; create an empty one
			err := os.MkdirAll(configPath, os.ModePerm)
			if err != nil {
				return
			}
			// Create an empty config file without prompting
			err = viper.SafeWriteConfig()
			if err != nil {
				return
			}
		}
	}
}
