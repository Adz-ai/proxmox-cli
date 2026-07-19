package utility

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newConfirmTestCmd(input string) (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "test"}
	AddYesFlag(cmd)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(input))
	return cmd, &out
}

func TestConfirmActionAcceptsYesInput(t *testing.T) {
	for _, input := range []string{"y\n", "yes\n", " Y \n"} {
		cmd, out := newConfirmTestCmd(input)
		if err := ConfirmAction(cmd, "Do the thing?"); err != nil {
			t.Errorf("input %q: unexpected error %v", input, err)
		}
		if !strings.Contains(out.String(), "Do the thing? [y/N]:") {
			t.Errorf("input %q: prompt not shown:\n%s", input, out.String())
		}
	}
}

func TestConfirmActionDeclines(t *testing.T) {
	for _, input := range []string{"n\n", "\n", ""} {
		cmd, _ := newConfirmTestCmd(input)
		err := ConfirmAction(cmd, "Do the thing?")
		if err == nil || !strings.Contains(err.Error(), "aborted") {
			t.Errorf("input %q: expected aborted error, got %v", input, err)
		}
	}
}

func TestConfirmActionSkippedWithYesFlag(t *testing.T) {
	cmd, out := newConfirmTestCmd("")
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	if err := ConfirmAction(cmd, "Do the thing?"); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("no prompt expected with --yes, got:\n%s", out.String())
	}
}
