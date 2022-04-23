package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-graphql/graph/generated"
	"zntr.io/hexagonal-bazel/cmd/urlshortener-graphql/graph/model"
)

func (r *mutationResolver) ShortenURL(ctx context.Context, url string, opts *model.ShortenInput) (*urlshortenerv1.Link, error) {
	req := &urlshortenerv1.CreateRequest{
		Url: url,
	}
	if opts != nil {
		req.Secret = opts.Secret
	}

	// Call datastore
	res, err := r.ShortenerAClient.Create(ctx, req)
	switch {
	case err != nil:
		return nil, err
	case res == nil:
		return nil, errors.New("unexpected nil response")
	case res.Error != nil:
		return nil, fmt.Errorf("%d - %s", res.Error.ErrorCode, res.Error.ErrorMessage)
	default:
	}

	// No error
	return res.Link, nil
}

func (r *mutationResolver) ResolveSecuredURL(ctx context.Context, id string, secret string) (*urlshortenerv1.Link, error) {
	req := &urlshortenerv1.ResolveRequest{
		Id:     id,
		Secret: &secret,
	}

	// Call datastore
	res, err := r.ShortenerAClient.Resolve(ctx, req)
	switch {
	case err != nil:
		return nil, err
	case res == nil:
		return nil, errors.New("unexpected nil response")
	case res.Error != nil:
		return nil, fmt.Errorf("%d - %s", res.Error.ErrorCode, res.Error.ErrorMessage)
	default:
	}

	// No error
	return res.Link, nil
}

func (r *queryResolver) Link(ctx context.Context, id string) (*urlshortenerv1.Link, error) {
	req := &urlshortenerv1.ResolveRequest{
		Id: id,
	}

	// Call datastore
	res, err := r.ShortenerAClient.Resolve(ctx, req)
	switch {
	case err != nil:
		return nil, err
	case res == nil:
		return nil, errors.New("unexpected nil response")
	case res.Error != nil:
		return nil, fmt.Errorf("%d - %s", res.Error.ErrorCode, res.Error.ErrorMessage)
	default:
	}

	// No error
	return res.Link, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
