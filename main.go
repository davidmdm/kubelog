package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/davidmdm/kubelog/internal/cmd"
	"github.com/davidmdm/kubelog/internal/cmd/list"
	"github.com/davidmdm/kubelog/internal/cmd/tail"

	"github.com/davidmdm/kubelog/internal/terminal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		stop()
	}()

	root := cmd.New()
	root.AddCommand(list.Cmd())
	root.AddCommand(tail.Cmd())

	if err := root.ExecuteContext(ctx); err != nil {
		terminal.PrintErrln()
		terminal.PrintErrln(err)
		os.Exit(1)
	}
}
