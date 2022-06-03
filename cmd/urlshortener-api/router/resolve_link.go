package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

// ResolveLink handles GET /links/{id}
func ResolveLink(shortener urlshortenerv1.ShortenerServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported operation for POST/GET only endpoint", http.StatusMethodNotAllowed)
			return
		}

		// Prepare request
		req := &urlshortenerv1.ResolveRequest{
			Id: chi.URLParam(r, "id"),
		}

		// Delegate to handler
		delegateTo(r.Context(), w, req, shortener.Resolve)
	}
}

// ResolveSecretLink handles POST /links/{id}
func ResolveSecretLink(shortener urlshortenerv1.ShortenerServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported operation for POST/GET only endpoint", http.StatusMethodNotAllowed)
			return
		}

		var req urlshortenerv1.ResolveRequest

		// Prepare json decoder
		dec := json.NewDecoder(io.LimitReader(r.Body, 512*1024))
		dec.DisallowUnknownFields()
		defer r.Body.Close()

		// Decode body
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "Unable to decode payload", http.StatusBadRequest)
			return
		}

		// Use urlparam for ID
		req.Id = chi.URLParam(r, "id")

		// Delegate to handler
		delegateTo(r.Context(), w, &req, shortener.Resolve)
	}
}
