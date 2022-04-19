package main

import (
	"errors"
	"log"
	"net"

	"github.com/timshannon/badgerhold/v4"
	"google.golang.org/grpc"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-datastore/server"
	"zntr.io/hexagonal-bazel/infrastructure/broker"
	"zntr.io/hexagonal-bazel/infrastructure/datastore/badger"
)

func main() {
	// Create the listener
	lis, err := net.Listen("tcp", "localhost:3001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create data repository
	options := badgerhold.DefaultOptions
	options.Dir = "data"
	options.ValueDir = "data"

	store, err := badgerhold.Open(options)
	defer store.Close()
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	urlshortenerv1.RegisterShortenerAPIServer(grpcServer,
		server.New(
			badger.Links(store),
			broker.Console(),
		),
	)

	// Serve
	if err := grpcServer.Serve(lis); err != nil {
		switch {
		case errors.Is(err, grpc.ErrServerStopped):
			return
		default:
			log.Fatal(err)
		}
	}
}
