package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/thesayyn/no/pkg/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := cmd.Root.ExecuteContext(ctx); err != nil {
		log.Fatal("error during command execution:", err)
	}
}
