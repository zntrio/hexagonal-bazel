package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

// ResolveLink handles GET /links/{id}
func ResolveLink(shortener urlshortenerv1.ShortenerAPIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported operation for GET only endpoint", http.StatusMethodNotAllowed)
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
