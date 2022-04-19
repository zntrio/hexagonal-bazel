package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-api/router"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Connect to datastore
	conn, err := grpc.Dial("localhost:3001", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create grpc client
	dsClient := urlshortenerv1.NewShortenerAPIClient(conn)

	// Default route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("cloud.urlshortener.v1"))
	})

	// API routes
	r.Route("/api/v1", func(sr chi.Router) {
		sr.Post("/links", router.CreateLink(dsClient))
		sr.Get("/links/{id}", router.ResolveLink(dsClient))
	})

	http.ListenAndServe(":3000", r)
}
