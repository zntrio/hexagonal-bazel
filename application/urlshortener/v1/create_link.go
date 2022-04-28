package v1

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
	"unicode/utf8"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/clock"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/reactor"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/infrastructure/serr"
	"zntr.io/hexagonal-bazel/pkg/types"
)

// Type aliases
type (
	CreateRequest     = urlshortener.CreateRequest
	CreateResponse    = urlshortener.CreateResponse
	CreateHandlerFunc = reactor.Handler[CreateRequest, CreateResponse]
)

const (
	// maxURLLength defines the max URL length authorized to shorten.
	maxURLLength = 2000
	// expireLowerBound defines the minimum allowed for link expiration duration in seconds.
	expireLowerBound = 30
)

// CreateHandler handles the urlshortener.Create request.
func CreateHandler(links link.Repository, codeGenerator generator.Generator[string], secretEncoder password.Hasher, secretGenerator generator.Generator[string], clockProvider clock.Clock) CreateHandlerFunc {
	return func(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
		var res CreateResponse

		// Check arguments
		if req == nil {
			err := errors.New("link: unable to process nil request")
			res.Error = serr.ServerError(err).Build()
			return &res, err
		}

		// Validate request
		if err := validation.ValidateStruct(req,
			// URL is mandatory and must contain a valid URL syntax.
			validation.Field(&req.Url, validation.Required, is.URL),
		); err != nil {
			res.Error = serr.InvalidRequest().Build(
				serr.InternalErr(err),
			)
			return &res, fmt.Errorf("link: unable to validate the request: %w", err)
		}

		// Check URL length
		if utf8.RuneCountInString(req.Url) > maxURLLength {
			res.Error = serr.InvalidRequest().Build(
				serr.Description("The given URL is too long (> 2000 characters)."),
				serr.Fields("url"),
			)
			return &res, errors.New("link: url is too long")
		}

		// Parse URL to normalize it
		u, err := url.Parse(req.Url)
		if err != nil {
			res.Error = serr.InvalidRequest().Build(
				serr.Fields("url"),
			)
			return &res, fmt.Errorf("link: unable to validate input url %q: %w", req.Url, err)
		}
		if u.Scheme == "" {
			// Default to https for empty scheme
			u.Scheme = "https"
		}

		// Skip operation if in validation mode
		if req.ValidateOnly {
			res.Link = &urlshortener.Link{
				Url: types.AsRef(u.String()),
			}
			return &res, nil
		}

		// Create a public identifier
		code, err := codeGenerator.Generate()
		if err != nil {
			res.Error = serr.ServerError(err).Build()
			return &res, fmt.Errorf("link: unable to generate public identifier: %w", err)
		}

		// Set required properties
		dopts := []link.DomainOption{
			// Generate a new identifier
			link.WithID(link.ID(code)),
			// Use parsed URL to normalize it
			link.WithURL(u.String()),
			// Set creation date
			link.WithCreatedAt(clockProvider.Now()),
		}

		// Secret required?
		if req.SecretRequired {
			// Generate a secret
			secret, err := secretGenerator.Generate()
			if err != nil {
				res.Error = serr.ServerError(err).Build()
				return &res, fmt.Errorf("link: unable to generate secret: %w", err)
			}

			// Derive secret hash from the secret value.
			sh, err := secretEncoder.Hash(secret)
			if err != nil {
				res.Error = serr.ServerError(err).Build()
				return &res, fmt.Errorf("link: unable to compute secret hash: %w", err)
			}

			// Add the secret hash.
			dopts = append(dopts, link.WithSecretHash(sh))

			// Return generated secret
			res.Secret = types.AsRef(secret)
		}

		// Expirable link
		if req.ExpiresIn != nil {
			// Validate expiration
			if *req.ExpiresIn < expireLowerBound {
				res.Error = serr.InvalidRequest().Build(
					serr.Description("The given expiration period is too short. It should be 30s at least."),
					serr.Fields("expires_in"),
				)
				return &res, fmt.Errorf("link: expiration too short for %q: %w", req.Url, err)
			}

			// Compute expiration date
			expiresAt := clockProvider.Now().Add(time.Duration(*req.ExpiresIn) * time.Second)

			// Add the expiration date.
			dopts = append(dopts, link.WithExpiresAt(expiresAt))
		}

		// Create a new Link.
		domainObject := link.New(dopts...)

		// Save to persistence
		if err := links.Save(ctx, domainObject); err != nil {
			res.Secret = nil
			res.Error = serr.ServerError(err).Build()
			return &res, fmt.Errorf("link: unable to create the shortened URL: %w", err)
		}

		// Prepare response
		res.Link = fromLink(domainObject)

		// No error
		return &res, nil
	}
}
