// Command journal is a terminal journaling tool: it opens dated markdown entries
// in your editor and records them as git commits.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/manningc33/journal/internal/cli"
	"github.com/manningc33/journal/internal/run"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := cli.App{
		Out:        os.Stdout,
		Err:        os.Stderr,
		Runner:     run.OS{},
		ConfigPath: os.Getenv("JOURNAL_CONFIG"),
	}
	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
