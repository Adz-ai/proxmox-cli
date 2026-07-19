package vm

import (
	"context"
	"fmt"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

type vmTaskSpec struct {
	use    string
	short  string
	long   string
	verb   string // present tense for errors, e.g. "start"
	done   string // past tense for the success message, e.g. "started"
	action func(context.Context, interfaces.VirtualMachineInterface) (*proxmox.Task, error)
}

func newVMTaskCmd(spec vmTaskSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Long:  spec.long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			vm, id, err := vmFromFlags(cmd)
			if err != nil {
				return err
			}

			task, err := spec.action(ctx, vm)
			if err != nil {
				return fmt.Errorf("%s VM %d: %w", spec.verb, id, err)
			}
			if err := utility.WaitForTask(ctx, task, utility.TaskTimeout(cmd)); err != nil {
				return fmt.Errorf("%s VM %d: %w", spec.verb, id, err)
			}

			fmt.Fprintf(out, "VM %d %s successfully\n", id, spec.done)
			return nil
		},
	}

	addVMTargetFlags(cmd)
	return cmd
}

func newStartCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "start",
		short: "Start a virtual machine",
		long:  `Start a stopped virtual machine on the specified node.`,
		verb:  "start",
		done:  "started",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Start(ctx)
		},
	})
}

func newStopCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "stop",
		short: "Stop a virtual machine immediately",
		long:  `Hard-stop a running virtual machine on the specified node. Use shutdown for a clean, guest-initiated stop.`,
		verb:  "stop",
		done:  "stopped",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Stop(ctx)
		},
	})
}

func newRestartCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "restart",
		short: "Restart a virtual machine",
		long:  `Reboot a running virtual machine on the specified node.`,
		verb:  "restart",
		done:  "restarted",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Reboot(ctx)
		},
	})
}

func newShutdownCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "shutdown",
		short: "Shut down a virtual machine gracefully",
		long:  `Request a clean, guest-initiated shutdown of a running virtual machine.`,
		verb:  "shut down",
		done:  "shut down",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Shutdown(ctx)
		},
	})
}

func newSuspendCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "suspend",
		short: "Suspend a virtual machine",
		long:  `Pause a running virtual machine, keeping its state in memory.`,
		verb:  "suspend",
		done:  "suspended",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Pause(ctx)
		},
	})
}

func newResumeCmd() *cobra.Command {
	return newVMTaskCmd(vmTaskSpec{
		use:   "resume",
		short: "Resume a suspended virtual machine",
		long:  `Resume a previously suspended virtual machine.`,
		verb:  "resume",
		done:  "resumed",
		action: func(ctx context.Context, vm interfaces.VirtualMachineInterface) (*proxmox.Task, error) {
			return vm.Resume(ctx)
		},
	})
}
