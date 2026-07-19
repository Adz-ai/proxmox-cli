package utility

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// AddYesFlag registers the shared --yes flag on a destructive command.
func AddYesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
}

// ConfirmAction asks the user to confirm a destructive action unless --yes
// was given. Declining, or running non-interactively without --yes, returns
// an error so scripts must opt in explicitly.
func ConfirmAction(cmd *cobra.Command, warning string) error {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return fmt.Errorf("read yes flag: %w", err)
	}
	if yes {
		return nil
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s [y/N]: ", warning)

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return errors.New("aborted; pass --yes to skip this prompt")
	}
}
