package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-api/router"
)

var (
	//go:embed static/*.html
	templateFS embed.FS
)

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	fmt.Fprintln(w, "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQEAYAAABPYyMiAAAABmJLR0T///////8JWPfcAAAACXBIWXMAAABIAAAASABGyWs+AAAAF0lEQVRIx2NgGAWjYBSMglEwCkbBSAcACBAAAeaR9cIAAAAASUVORK5CYII=")
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Prepare all templates
	templates, err := template.ParseFS(templateFS, "static/*.html")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to datastore
	conn, err := grpc.Dial("localhost:3001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create grpc client
	dsClient := urlshortenerv1.NewShortenerAPIClient(conn)

	// Set default handlers
	r.NotFound(router.NotFound(templates))
	r.MethodNotAllowed(router.NotFound(templates))

	// Favico
	r.HandleFunc("/favicon.ico", favicon)

	// API routes
	r.Route("/api/v1", func(sr chi.Router) {
		sr.Post("/links", router.CreateLink(dsClient))
		sr.Get("/links/{id}", router.ResolveLink(dsClient))
		sr.Post("/links/{id}", router.ResolveSecretLink(dsClient))
	})

	// Index
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			panic(err)
		}
	})

	// Catch-all handler
	r.Handle("/*", router.Home(dsClient, templates))

	http.ListenAndServe(":3000", r)
}
