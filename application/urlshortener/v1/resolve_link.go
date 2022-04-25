package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/pkg/reactor"
)

// Type aliases
type (
	ResolveRequest  = urlshortener.ResolveRequest
	ResolveResponse = urlshortener.ResolveResponse
)

// ResolveHandler handles the urlshortener.Resolve request.
func ResolveHandler(links link.Resolver, secretVerifier password.Verifier) reactor.Handler[ResolveRequest, ResolveResponse] {
	return func(ctx context.Context, req *ResolveRequest) (*ResolveResponse, error) {
		var res ResolveResponse

		// Check arguments
		if req == nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to process nil request.",
				ErrorCode:    http.StatusBadRequest,
			}
			return &res, nil
		}

		// Validate request
		if err := validation.ValidateStruct(req,
			// ID is mandatory and must contain a valid ascii string.
			validation.Field(&req.Id, validation.Required, is.ASCII),
		); err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to validate request execution conditions.",
				ErrorCode:    http.StatusPreconditionFailed,
			}
			return &res, nil
		}

		// Save to persistence
		m, err := links.GetByID(ctx, link.ID(req.Id))
		switch {
		case err == nil:
			// Skip processing
		case errors.Is(err, link.ErrLinkNotFound):
			res.Error = &errorsv1.Error{
				ErrorMessage: "Link not found.",
				ErrorCode:    http.StatusNotFound,
			}
			return &res, nil
		case err != nil:
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to process the request.",
				ErrorCode:    http.StatusInternalServerError,
			}
			return &res, fmt.Errorf("unable to resolve %q: %w", req.Id, err)
		}

		// Check if secret is required to reveal the url
		if m.GetSecretHash() != "" {
			if req.Secret == nil {
				res.Error = &errorsv1.Error{
					ErrorMessage: "This shortened url requires a secret to be revealed.",
					ErrorCode:    http.StatusNotAcceptable,
				}
				return &res, nil
			}

			// Verify the secret match
			if err := secretVerifier.Verify(*req.Secret, m.GetSecretHash()); err != nil {
				res.Error = &errorsv1.Error{
					ErrorMessage: "This shortened url requires a valid secret to be revealed.",
					ErrorCode:    http.StatusForbidden,
				}
				return &res, nil
			}
		}

		// Prepare response
		res.Link = fromLink(m)

		// No error
		return &res, nil
	}
}
