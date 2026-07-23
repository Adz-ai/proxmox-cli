package utility

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// RunConsole pumps a termproxy websocket to and from the local terminal
// until the user presses Ctrl+] or the connection closes. The caller's
// closer is always invoked on return.
func RunConsole(cmd *cobra.Command, send, recv chan []byte, errs chan error, closer func() error) error {
	return RunConsoleStreams(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), send, recv, errs, closer)
}

// RunConsoleStreams is RunConsole against explicit streams so callers other
// than cobra commands (the TUI) can attach a console. stdin must be an
// interactive terminal.
func RunConsoleStreams(ctx context.Context, stdin io.Reader, out io.Writer, send, recv chan []byte, errs chan error, closer func() error) error {
	defer func() { _ = closer() }()

	file, ok := stdin.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return errors.New("console requires an interactive terminal")
	}

	oldState, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return fmt.Errorf("switch terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(file.Fd()), oldState) }()

	fmt.Fprint(out, "Connected. Press Ctrl+] to disconnect.\r\n")

	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 1024)
		for {
			n, err := file.Read(buffer)
			if err != nil {
				return
			}
			data := make([]byte, n)
			copy(data, buffer[:n])
			for _, char := range data {
				if char == 0x1d { // Ctrl+]
					return
				}
			}
			select {
			case send <- data:
			case <-done:
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			fmt.Fprint(out, "\r\nDisconnected\r\n")
			return nil
		case data, open := <-recv:
			if !open {
				fmt.Fprint(out, "\r\nConnection closed\r\n")
				return nil
			}
			if _, err := out.Write(data); err != nil {
				return err
			}
		case err := <-errs:
			if err != nil {
				return fmt.Errorf("console connection error: %w", err)
			}
		}
	}
}
