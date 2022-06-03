package actor

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
	"github.com/timshannon/badgerhold/v4"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-datastore/actor/server"
	"zntr.io/hexagonal-bazel/infrastructure/datastore/badger"
	"zntr.io/hexagonal-bazel/infrastructure/generator/passphrase"
	"zntr.io/hexagonal-bazel/infrastructure/generator/shortid"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
)

type dataStoreActor struct {
	cfg        *Configuration
	grpcServer *grpc.Server
	store      *badgerhold.Store
}

func Datastore(cfg *Configuration) Actor {
	return &dataStoreActor{
		cfg: cfg,
	}
}

// -----------------------------------------------------------------------------

func (a *dataStoreActor) Run(ctx context.Context) error {
	// Create actor execution context.
	actorCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create data repository
	badgerOpts := badgerhold.DefaultOptions
	badgerOpts.Dir = a.cfg.Persistence.Badger.Dir
	badgerOpts.ValueDir = a.cfg.Persistence.Badger.ValueDir

	// Start persistence
	store, err := badgerhold.Open(badgerOpts)
	if err != nil {
		return fmt.Errorf("unable to open data store: %w", err)
	}

	// Assign to actor
	a.store = store

	// Initialize gRPC server
	a.grpcServer = grpc.NewServer()
	urlshortenerv1.RegisterShortenerServiceServer(a.grpcServer,
		server.New(
			badger.Links(store),
			shortid.New(a.cfg.Generator.WorkerID),
			password.Argon2(),
			passphrase.Diceware(a.cfg.Generator.SecretWordCount),
		),
	)

	eg, _ := errgroup.WithContext(ctx)

	// Close signal handler
	eg.Go(func() error {
		<-actorCtx.Done()
		return actorCtx.Err()
	})

	// Serve grpc
	eg.Go(func() error {
		// Create the listener
		lis, err := net.Listen(a.cfg.Server.Listener.Network, a.cfg.Server.Listener.Address)
		if err != nil {
			return fmt.Errorf("unable to initialize server listener: %w", err)
		}

		// Start the service
		return a.grpcServer.Serve(lis)
	})

	return eg.Wait()
}

func (a *dataStoreActor) Close() error {
	var result error

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.store != nil {
		result = multierror.Append(result, a.store.Close())
	}

	return result
}
