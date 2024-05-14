/*
Copyright © 2024 Adarssh Athithan
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proxmox-cli",
	Short: "proxmox-cli is a CLI for Proxmox management",
	Long:  `proxmox-cli is a Command Line Interface built using Cobra for managing Proxmox servers.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
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

	// Attempt to read the config, if it doesn't exist, create it with default settings
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// The Config file does not exist; create it with some default values
			err := os.MkdirAll(configPath, os.ModePerm)
			if err != nil {
				return
			}
			setupInitialConfig()
			err = viper.SafeWriteConfig()
			if err != nil {
				return
			}
		}
	}
}

func setupInitialConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Proxmox server URL: ")
	serverURL, _ := reader.ReadString('\n')
	serverURL = strings.TrimSpace(serverURL)

	// Store the server URL in the configuration file
	viper.Set("server_url", serverURL)
	fmt.Println("Configuration saved. You can now use the CLI.")
}
