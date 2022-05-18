package router

import (
	"html/template"
	"net/http"

	"zntr.io/hexagonal-bazel/infrastructure/serr"
)

func NotFound(templates *template.Template) http.HandlerFunc {
	defaultError := serr.ResourceNotFound().Build()

	return func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "error.html", defaultError); err != nil {
			panic(err)
		}
	}
}
