package graph

import urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	ShortenerAClient urlshortenerv1.ShortenerServiceClient
}
