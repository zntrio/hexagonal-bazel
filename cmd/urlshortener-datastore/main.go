package main

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-datastore/actor"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)

	// Prepare the actor
	dataStore := actor.Datastore(actor.DefaultConfiguration())

	// Start the datastore actor
	g.Go(func() error { return dataStore.Run(gctx) })

	// Graceful stop goroutine
	g.Go(func() error {
		<-gctx.Done()
		stop()
		return dataStore.Close()
	})

	// Wait for all errgroup goroutines
	err := g.Wait()
	switch {
	case err == nil, errors.Is(err, context.Canceled):
		fmt.Println("finished clean")
	case err != nil:
		fmt.Printf("received error: %v", err)
	}
}
