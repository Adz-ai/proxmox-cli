package cmd

import (
	"bufio"
	"fmt"
	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize or reconfigure Proxmox CLI",
		Long: `Initialize the Proxmox CLI configuration.
	
This command helps you set up or reconfigure your connection to a Proxmox VE server.
It will prompt you for the server URL and save it to the configuration file.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			in := cmd.InOrStdin()
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			insecure, err := cmd.Flags().GetBool("insecure")
			if err != nil {
				return err
			}
			caCert, err := cmd.Flags().GetString("ca-cert")
			if err != nil {
				return err
			}
			caCert = strings.TrimSpace(caCert)

			// Check if already configured
			existingURL := viper.GetString("server_url")
			if existingURL != "" && !force {
				fmt.Fprintf(out, "Already configured for server: %s\n", existingURL)
				fmt.Fprintln(out, "Use --force to reconfigure")
				return nil
			}
			existingInsecure := viper.GetBool("insecure")
			existingCACert := viper.GetString("ca_cert")
			if existingURL != "" && !cmd.Flags().Changed("insecure") {
				insecure = existingInsecure
			}
			if existingURL != "" && !cmd.Flags().Changed("ca-cert") {
				caCert = existingCACert
			}
			if insecure && caCert != "" {
				return fmt.Errorf("--insecure and --ca-cert cannot be used together")
			}

			reader := bufio.NewReader(in)

			fmt.Fprintln(out, "Welcome to Proxmox CLI")
			fmt.Fprintln(out, "Proxmox CLI Configuration")
			fmt.Fprintln(out, "=========================")

			if existingURL != "" {
				fmt.Fprintf(out, "Current server: %s\n\n", existingURL)
			}

			fmt.Fprint(out, "Enter Proxmox server URL (e.g., https://192.168.1.100:8006): ")

			// For testing, we need to handle the case where stdin is not a terminal
			serverURL, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return fmt.Errorf("read server URL: %w", err)
			}
			serverURL, err = utility.NormalizeServerURL(serverURL)
			if err != nil {
				return err
			}

			viper.Set("server_url", serverURL)
			viper.Set("insecure", insecure)
			viper.Set("ca_cert", caCert)

			if existingURL != "" && (existingURL != serverURL || existingInsecure != insecure || existingCACert != caCert) {
				utility.ClearAuthTicket()
				fmt.Fprintln(out, "Cleared existing authentication (connection settings changed)")
			}

			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("save configuration: %w", err)
			}
			configPath, err := utility.ConfigFile()
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "\nConfiguration saved to %s\n", configPath)
			fmt.Fprintf(out, "Server URL: %s\n", serverURL)
			fmt.Fprintln(out, "\nNext step: Run 'proxmox-cli auth login -u <username>' to authenticate")
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force reconfiguration even if already configured")
	cmd.Flags().Bool("insecure", false, "Skip TLS certificate verification")
	cmd.Flags().String("ca-cert", "", "Path to a custom CA certificate")
	return cmd
}
