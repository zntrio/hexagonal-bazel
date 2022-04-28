package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/clock"
	"zntr.io/hexagonal-bazel/infrastructure/reactor"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/infrastructure/serr"
)

// Type aliases
type (
	ResolveRequest     = urlshortener.ResolveRequest
	ResolveResponse    = urlshortener.ResolveResponse
	ResolveHandlerFunc reactor.Handler[ResolveRequest, ResolveResponse]
)

// ResolveHandler handles the urlshortener.Resolve request.
func ResolveHandler(links link.Resolver, secretVerifier password.Verifier, clockProvider clock.Clock) ResolveHandlerFunc {
	return func(ctx context.Context, req *ResolveRequest) (*ResolveResponse, error) {
		var res ResolveResponse

		// Check arguments
		if req == nil {
			err := errors.New("link: unable to process nil request")
			res.Error = serr.ServerError(err).Build()
			return &res, err
		}

		// Validate request
		if err := validation.ValidateStruct(req,
			// ID is mandatory and must contain a valid ascii string.
			validation.Field(&req.Id, validation.Required, is.ASCII),
		); err != nil {
			res.Error = serr.InvalidRequest().Build(
				serr.InternalErr(err),
			)
			return &res, fmt.Errorf("link: unable to validate the request: %w", err)
		}

		// Save to persistence
		m, err := links.GetByID(ctx, link.ID(req.Id))
		switch {
		case err == nil:
			// Skip processing
		case errors.Is(err, link.ErrLinkNotFound):
			res.Error = serr.ResourceNotFound().Build()
			return &res, fmt.Errorf("link: link %q not found: %w", req.Id, err)
		case err != nil:
			res.Error = serr.ServerError(err).Build()
			return &res, fmt.Errorf("link: unable to resolve %q: %w", req.Id, err)
		}

		// Check if secret is expired.
		if m.IsExpired(clockProvider.Now()) {
			res.Error = serr.ResourceNotFound().Build()
			return &res, fmt.Errorf("link: expiration reached for %q", m.GetID())
		}

		// Check if secret is required to reveal the url
		if m.IsProtected() {
			if req.Secret == nil {
				res.Error = serr.InvalidRequest().Build(
					serr.Description("This shortened url requires a secret to be revealed."),
					serr.StatusCode(http.StatusNotAcceptable),
				)
				return &res, fmt.Errorf("link: secret required for %q", m.GetID())
			}

			// Verify the secret match
			if err := secretVerifier.Verify(*req.Secret, m.GetSecretHash()); err != nil {
				res.Error = serr.AccessDenied().Build(
					serr.Description("This shortened url requires a valid secret to be revealed."),
				)
				return &res, fmt.Errorf("link: invalid secret for %q", m.GetID())
			}
		}

		// Prepare response
		res.Link = fromLink(m)

		// No error
		return &res, nil
	}
}
