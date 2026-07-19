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
			if !utility.HasSessionTicket() {
				return fmt.Errorf("console requires a session ticket; run 'proxmox-cli auth login -u <username>' (API tokens cannot open websockets)")
			}

			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
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
