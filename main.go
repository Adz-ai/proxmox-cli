package main

import (
	"context"
	"github.com/Adz-ai/proxmox-cli/cmd"
	"os"
	"os/signal"
	"syscall"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cmd.NewRootCmd()
	root.Version = version
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
