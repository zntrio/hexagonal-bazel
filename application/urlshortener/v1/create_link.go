package v1

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gosimple/slug"

	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/clock"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/generator/passphrase"
	"zntr.io/hexagonal-bazel/infrastructure/reactor"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/infrastructure/serr"
	"zntr.io/hexagonal-bazel/pkg/types"
)

// Type aliases
type (
	CreateRequest     = urlshortenerv1.CreateRequest
	CreateResponse    = urlshortenerv1.CreateResponse
	CreateHandlerFunc = reactor.Handler[CreateRequest, CreateResponse]
)

const (
	// maxURLLength defines the max URL length authorized to shorten.
	maxURLLength = 2000
	// expireLowerBound defines the minimum allowed for link expiration duration in seconds.
	expireLowerBound = 30 // 30s
	// expireUpperBound defines the maximum allowed for link expiration duration in seconds.
	expireUpperBound = 7 * 24 * 60 * 60 // 7 days
)

// CreateHandler handles the urlshortener.Create request.
func CreateHandler(links link.Repository, codeGenerator generator.Generator[string], secretEncoder password.Hasher, secretGenerator passphrase.Generator, clockProvider clock.Clock) CreateHandlerFunc {
	return func(ctx context.Context, req *urlshortenerv1.CreateRequest) (*urlshortenerv1.CreateResponse, error) {
		res := urlshortenerv1.CreateResponse{}

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
			// Slug must be ascii only
			validation.Field(&req.Slug, is.PrintableASCII, validation.Length(5, 50)),
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
			res.Link = &urlshortenerv1.Link{
				Url: types.AsRef(u.String()),
			}
			return &res, nil
		}

		// Set required properties
		dopts := []link.DomainOption{
			// Use parsed URL to normalize it
			link.WithURL(u.String()),
			// Set creation date
			link.WithCreatedAt(clockProvider.Now()),
		}

		linkID := ""

		// Has specific slug
		if req.Slug != nil {
			// Set slug value
			linkID = strings.ToLower(strings.TrimSpace(*req.Slug))

			// Check slug syntax
			if !slug.IsSlug(linkID) {
				res.Error = serr.InvalidRequest().Build(
					serr.Descriptionf("The given slug value %q has not a valid syntax.", linkID),
					serr.Fields("slug"),
				)
				return &res, fmt.Errorf("link: invalid slug identifier: %w", err)
			}

			// Ensure link ID uniquenss
			_, err = links.GetByID(ctx, link.ID(linkID))
			switch {
			case errors.Is(err, link.ErrLinkNotFound):
				// Expected
			case err != nil:
				res.Error = serr.ServerError(err).Build()
				return &res, fmt.Errorf("link: unable to ensure link identifier uniqueness: %w", err)
			case err == nil:
				res.Error = serr.InvalidRequest().Build(
					serr.Description("The slug must be unique."),
					serr.Fields("slug"),
				)
				return &res, fmt.Errorf("link: unable to ensure link identifier uniqueness %q", linkID)
			}
		} else {
			// Create a public identifier
			code, err := codeGenerator.Generate()
			if err != nil {
				res.Error = serr.ServerError(err).Build()
				return &res, fmt.Errorf("link: unable to generate public identifier: %w", err)
			}

			// Set generated link identifier
			linkID = code
		}

		// Add link identifier.
		dopts = append(dopts, link.WithID(link.ID(linkID)))

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
			if *req.ExpiresIn < expireLowerBound || *req.ExpiresIn > expireUpperBound {
				res.Error = serr.InvalidRequest().Build(
					serr.Descriptionf(
						"The given expiration period is too short. It should be %d as minimum and %d as maximum value.",
						expireLowerBound, expireUpperBound,
					),
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

		// Overwrite link identifier with the given one
		res.Link.Id = linkID

		// No error
		return &res, nil
	}
}
