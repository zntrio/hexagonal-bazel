package router

import (
	"html/template"
	"net/http"
	"strings"

	"google.golang.org/grpc/status"
	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

// Home handles link resolution from the browser
func Home(shortener urlshortenerv1.ShortenerServiceClient, templates *template.Template) http.HandlerFunc {
	var displayError = func(w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			st := status.Convert(err)
			for _, detail := range st.Details() {
				switch t := detail.(type) {
				case *errorsv1.Error:
					if err := templates.ExecuteTemplate(w, "error.html", t); err != nil {
						panic(err)
					}
					return
				default:
					continue
				}
			}
		}
	}

	var doRedirect = func(w http.ResponseWriter, r *http.Request, l *urlshortenerv1.Link) {
		http.Redirect(w, r, *l.Url, http.StatusFound)
	}

	var doGet = func(w http.ResponseWriter, r *http.Request) {
		// Decode path
		var (
			ctx     = r.Context()
			shortID = strings.TrimPrefix(r.URL.Path, "/")
		)

		// Try to resolve the link
		res, err := shortener.Resolve(ctx, &urlshortenerv1.ResolveRequest{
			Id: shortID,
		})
		switch {
		case err != nil, res == nil, res.Link == nil:
			displayError(w, r, err)
		default:
			doRedirect(w, r, res.Link)
		}
	}

	var doPost = func(w http.ResponseWriter, r *http.Request) {
		// Parse input form
		if err := r.ParseForm(); err != nil {
			displayError(w, r, err)
			return
		}

		// Unpack values
		var (
			ctx       = r.Context()
			idRaw     = r.Form.Get("id")
			secretRaw = r.Form.Get("secret")
		)

		// Try to resolve the link
		res, err := shortener.Resolve(ctx, &urlshortenerv1.ResolveRequest{
			Id:     idRaw,
			Secret: &secretRaw,
		})
		switch {
		case err != nil, res == nil, res.Link == nil:
			displayError(w, r, err)
		default:
			doRedirect(w, r, res.Link)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			doGet(w, r)
		case http.MethodPost:
			doPost(w, r)
		default:
			http.Error(w, "Unsupported operation for GET/POST only endpoint", http.StatusMethodNotAllowed)
			return
		}
	}
}
