package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiurlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	urlshortenerv1 "zntr.io/hexagonal-bazel/application/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/clock"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/generator/passphrase"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
)

func New(store link.Repository, codeGenerator generator.Generator[string], secretStrategy password.Strategy, passphraseGenerator passphrase.Generator) apiurlshortenerv1.ShortenerAPIServer {
	// No error
	return &urlShortenerServer{
		createHandler:  urlshortenerv1.CreateHandler(store, codeGenerator, secretStrategy, passphraseGenerator, clock.Real()),
		resolveHandler: urlshortenerv1.ResolveHandler(store, secretStrategy, clock.Real()),
	}
}

type urlShortenerServer struct {
	apiurlshortenerv1.UnimplementedShortenerAPIServer

	createHandler  urlshortenerv1.CreateHandlerFunc
	resolveHandler urlshortenerv1.ResolveHandlerFunc
}

// Create a shortened link from the given URL.
func (s *urlShortenerServer) Create(ctx context.Context, req *urlshortenerv1.CreateRequest) (*urlshortenerv1.CreateResponse, error) {
	res, err := s.createHandler(ctx, req)
	if err == nil {
		return res, nil
	}

	st := status.New(codes.FailedPrecondition, res.Error.ErrorCode)
	st, err = st.WithDetails(res.Error)
	if err != nil {
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	}

	return res, st.Err()
}

// Resolve the shortened URL.
func (s *urlShortenerServer) Resolve(ctx context.Context, req *urlshortenerv1.ResolveRequest) (*urlshortenerv1.ResolveResponse, error) {
	return s.resolveHandler(ctx, req)
}
