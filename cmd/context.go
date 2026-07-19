package cmd

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage cluster configuration contexts",
		Long: `Work with multiple Proxmox clusters from one CLI. Each context holds
its own server URL, TLS settings, and credentials.

Create a context by configuring it directly:

  proxmox-cli --context work init
  proxmox-cli --context work auth login -u root@pam

Then switch with 'context use', or target any context for a single
command with the global --context flag.`,
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newContextListCmd(), newContextUseCmd(), newContextDeleteCmd())
	return cmd
}

func newContextListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured contexts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			format, err := utility.OutputFormat(cmd)
			if err != nil {
				return err
			}

			contexts := utility.ListContexts()
			if format == "json" {
				return utility.PrintJSON(out, contexts)
			}

			fmt.Fprintf(out, "%-3s %-20s %s\n", "", "Name", "Server")
			fmt.Fprintf(out, "%-3s %-20s %s\n", "", "----", "------")
			for _, context := range contexts {
				marker := ""
				if context.Current {
					marker = "*"
				}
				fmt.Fprintf(out, "%-3s %-20s %s\n", marker, context.Name, context.ServerURL)
			}
			if len(contexts) == 0 {
				fmt.Fprintln(out, "No contexts configured; run 'proxmox-cli init' to create one")
			}
			return nil
		},
	}

	utility.AddOutputFlag(cmd)
	return cmd
}

func newContextUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch the current context",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			names := []string{}
			for _, context := range utility.ListContexts() {
				names = append(names, context.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := args[0]
			if err := utility.ValidateContextName(name); err != nil {
				return err
			}
			if !utility.ContextExists(name) {
				return fmt.Errorf("context %q does not exist; configure it with 'proxmox-cli --context %s init'", name, name)
			}

			viper.Set("current_context", name)
			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("save current context: %w", err)
			}

			fmt.Fprintf(out, "Switched to context %q\n", name)
			return nil
		},
	}
}

func newContextDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a context and its credentials",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			names := []string{}
			for _, context := range utility.ListContexts() {
				if !context.Current {
					names = append(names, context.Name)
				}
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := args[0]
			if err := utility.ValidateContextName(name); err != nil {
				return err
			}
			if !utility.ContextExists(name) {
				return fmt.Errorf("context %q does not exist", name)
			}
			if name == utility.ActiveContext() {
				return fmt.Errorf("cannot delete the active context; switch first with 'proxmox-cli context use <other>'")
			}

			utility.DeleteContext(name)
			if err := utility.WriteConfig(); err != nil {
				return fmt.Errorf("delete context: %w", err)
			}

			fmt.Fprintf(out, "Context %q deleted\n", name)
			return nil
		},
	}
}
