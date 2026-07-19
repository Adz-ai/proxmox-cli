package lxc

import (
	"context"
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

type containerTaskSpec struct {
	use    string
	short  string
	long   string
	verb   string // present tense for errors, e.g. "start"
	done   string // past tense for the success message, e.g. "started"
	flags  func(*cobra.Command)
	action func(*cobra.Command, context.Context, interfaces.ContainerInterface) (*proxmox.Task, error)
}

func newContainerTaskCmd(spec containerTaskSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Long:  spec.long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			container, vmid, err := containerFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := spec.action(cmd, ctx, container)
			if err != nil {
				return fmt.Errorf("%s container %d: %w", spec.verb, vmid, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd), out); err != nil {
				return fmt.Errorf("%s container %d: %w", spec.verb, vmid, err)
			}

			fmt.Fprintf(out, "Container %d %s successfully\n", vmid, spec.done)
			return nil
		},
	}

	addContainerTargetFlags(cmd)
	if spec.flags != nil {
		spec.flags(cmd)
	}
	return cmd
}

func newStartCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "start",
		short: "Start an LXC container",
		verb:  "start",
		done:  "started",
		action: func(_ *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			return container.Start(ctx)
		},
	})
}

func newStopCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "stop",
		short: "Stop an LXC container immediately",
		long:  `Hard-stop a running LXC container. Use shutdown for a clean stop.`,
		verb:  "stop",
		done:  "stopped",
		action: func(_ *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			return container.Stop(ctx)
		},
	})
}

func newRestartCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "restart",
		short: "Restart an LXC container",
		long:  `Reboot a running LXC container on the specified node.`,
		verb:  "restart",
		done:  "restarted",
		action: func(_ *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			return container.Reboot(ctx)
		},
	})
}

func newShutdownCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "shutdown",
		short: "Shut down an LXC container gracefully",
		long:  `Request a clean shutdown of a running LXC container.`,
		verb:  "shut down",
		done:  "shut down",
		flags: func(cmd *cobra.Command) {
			cmd.Flags().Bool("force", false, "Hard-stop the container if the clean shutdown times out")
			cmd.Flags().Int("grace-seconds", 60, "Seconds to wait for the clean shutdown before giving up")
		},
		action: func(cmd *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return nil, fmt.Errorf("read force flag: %w", err)
			}
			graceSeconds, err := cmd.Flags().GetInt("grace-seconds")
			if err != nil {
				return nil, fmt.Errorf("read grace-seconds flag: %w", err)
			}
			if graceSeconds <= 0 {
				return nil, fmt.Errorf("grace-seconds must be positive")
			}
			return container.Shutdown(ctx, force, graceSeconds)
		},
	})
}

func newSuspendCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "suspend",
		short: "Suspend an LXC container",
		verb:  "suspend",
		done:  "suspended",
		action: func(_ *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			return container.Suspend(ctx)
		},
	})
}

func newResumeCmd() *cobra.Command {
	return newContainerTaskCmd(containerTaskSpec{
		use:   "resume",
		short: "Resume a suspended LXC container",
		verb:  "resume",
		done:  "resumed",
		action: func(_ *cobra.Command, ctx context.Context, container interfaces.ContainerInterface) (*proxmox.Task, error) {
			return container.Resume(ctx)
		},
	})
}
