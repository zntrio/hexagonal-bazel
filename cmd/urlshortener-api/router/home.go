package router

import (
	"net/http"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

// Home handles link resolution from the browser
func Home(shortener urlshortenerv1.ShortenerAPIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		case http.MethodPost:
		default:
		}
	}
}
