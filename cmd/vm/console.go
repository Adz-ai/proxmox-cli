package vm

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open an interactive console to a virtual machine",
		Long: `Attach to the VM's terminal via the Proxmox terminal proxy.

Requires password (session) authentication; Proxmox does not allow API
tokens to open console websockets. Press Ctrl+] to disconnect.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			node, id, err := vmTargetFromFlags(cmd)
			if err != nil {
				return err
			}

			// The regular client prefers API-token auth, which Proxmox
			// rejects for console websockets; use the session-only client.
			client, err := utility.SessionClient()
			if err != nil {
				return err
			}
			retrievedNode, err := client.Node(ctx, node)
			if err != nil {
				return fmt.Errorf("get node %q: %w", node, err)
			}
			vm, err := retrievedNode.VirtualMachine(ctx, id)
			if err != nil {
				return fmt.Errorf("get VM %d: %w", id, err)
			}

			term, err := vm.TermProxy(ctx)
			if err != nil {
				return fmt.Errorf("open terminal proxy for VM %d: %w", id, err)
			}
			send, recv, errs, closer, err := vm.TermWebSocket(term)
			if err != nil {
				return fmt.Errorf("connect console websocket for VM %d: %w", id, err)
			}
			return utility.RunConsole(cmd, send, recv, errs, closer)
		},
	}

	addVMTargetFlags(cmd)
	return cmd
}
