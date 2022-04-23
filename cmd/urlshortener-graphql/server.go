package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-graphql/graph"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-graphql/graph/generated"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Connect to datastore
	conn, err := grpc.Dial("localhost:3001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create grpc client
	dsClient := urlshortenerv1.NewShortenerAPIClient(conn)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		ShortenerAClient: dsClient,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
