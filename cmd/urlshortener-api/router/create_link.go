package router

import (
	"encoding/json"
	"io"
	"net/http"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

// CreateLink handles POST /links
func CreateLink(shortener urlshortenerv1.ShortenerServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported operation for POST only endpoint", http.StatusMethodNotAllowed)
			return
		}

		var req urlshortenerv1.CreateRequest

		// Prepare json decoder
		dec := json.NewDecoder(io.LimitReader(r.Body, 512*1024))
		dec.DisallowUnknownFields()
		defer r.Body.Close()

		// Decode body
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "Unable to decode payload", http.StatusBadRequest)
			return
		}

		// Delegate to handler
		delegateTo(r.Context(), w, &req, shortener.Create)
	}
}
