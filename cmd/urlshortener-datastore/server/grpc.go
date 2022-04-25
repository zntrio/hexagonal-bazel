package server

import (
	"context"

	apiurlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	urlshortenerv1 "zntr.io/hexagonal-bazel/application/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
)

func New(store link.Repository, publisher eventbus.EventPublisher, codeGenerator generator.Generator[string], secretStrategy password.Strategy) apiurlshortenerv1.ShortenerAPIServer {
	// No error
	return &urlShortenerServer{
		createHandler:  urlshortenerv1.CreateHandler(store, publisher, codeGenerator, secretStrategy),
		resolveHandler: urlshortenerv1.ResolveHandler(store, secretStrategy),
	}
}

type urlShortenerServer struct {
	apiurlshortenerv1.UnimplementedShortenerAPIServer

	createHandler  func(context.Context, *urlshortenerv1.CreateRequest) (*urlshortenerv1.CreateResponse, error)
	resolveHandler func(context.Context, *urlshortenerv1.ResolveRequest) (*urlshortenerv1.ResolveResponse, error)
}

// Create a shortened link from the given URL.
func (s *urlShortenerServer) Create(ctx context.Context, req *urlshortenerv1.CreateRequest) (*urlshortenerv1.CreateResponse, error) {
	return s.createHandler(ctx, req)
}

// Resolve the shortened URL.
func (s *urlShortenerServer) Resolve(ctx context.Context, req *urlshortenerv1.ResolveRequest) (*urlshortenerv1.ResolveResponse, error) {
	return s.resolveHandler(ctx, req)
}
