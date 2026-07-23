package lxc

import (
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/spf13/cobra"
)

func newConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Open an interactive console to an LXC container",
		Long: `Attach to the container's terminal via the Proxmox terminal proxy.

Requires password (session) authentication; Proxmox does not allow API
tokens to open console websockets. Press Ctrl+] to disconnect.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			nodeName, vmid, err := containerTargetFromFlags(cmd)
			if err != nil {
				return err
			}

			// The regular client prefers API-token auth, which Proxmox
			// rejects for console websockets; use the session-only client.
			client, err := utility.SessionClient()
			if err != nil {
				return err
			}
			retrievedNode, err := client.Node(ctx, nodeName)
			if err != nil {
				return fmt.Errorf("get node %q: %w", nodeName, err)
			}
			container, err := retrievedNode.Container(ctx, vmid)
			if err != nil {
				return fmt.Errorf("get container %d: %w", vmid, err)
			}

			term, err := container.TermProxy(ctx)
			if err != nil {
				return fmt.Errorf("open terminal proxy for container %d: %w", vmid, err)
			}
			send, recv, errs, closer, err := container.TermWebSocket(term)
			if err != nil {
				return fmt.Errorf("connect console websocket for container %d: %w", vmid, err)
			}
			return utility.RunConsole(cmd, send, recv, errs, closer)
		},
	}

	addContainerTargetFlags(cmd)
	return cmd
}
