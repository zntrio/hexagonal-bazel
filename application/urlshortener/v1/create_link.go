package v1

import (
	"context"
	"net/http"
	"net/url"
	"unicode/utf8"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
	"zntr.io/hexagonal-bazel/pkg/reactor"
	"zntr.io/hexagonal-bazel/pkg/types"
)

// Type aliases
type (
	CreateRequest  = urlshortener.CreateRequest
	CreateResponse = urlshortener.CreateResponse
)

const (
	// maxURLLength defines the max URL length authorized to shorten.
	maxURLLength = 2000
)

// CreateHandler handles the urlshortener.Create request.
func CreateHandler(links link.Repository, publisher eventbus.EventPublisher, codeGenerator generator.Generator[string], secretEncoder password.Hasher) reactor.Handler[CreateRequest, CreateResponse] {
	return func(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
		var res CreateResponse

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
			// URL is mandatory and must contain a valid URL syntax.
			validation.Field(&req.Url, validation.Required, is.URL),
		); err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to validate request execution conditions.",
				ErrorCode:    http.StatusPreconditionFailed,
			}
			return &res, err
		}

		// Check URL length
		if utf8.RuneCountInString(req.Url) > maxURLLength {
			res.Error = &errorsv1.Error{
				ErrorMessage: "The given URL is too long (> 2000 characters).",
				ErrorCode:    http.StatusBadRequest,
			}
			return &res, nil
		}

		// Parse URL to normalize it
		u, err := url.Parse(req.Url)
		if err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to validate URL.",
				ErrorCode:    http.StatusBadRequest,
			}
			return &res, err
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
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to generate shortened identifier.",
				ErrorCode:    http.StatusInternalServerError,
			}
			return &res, err
		}

		// Set required properties
		dopts := []link.DomainOption{
			// Generate a new identifier
			link.WithID(link.ID(code)),
			// Use parsed URL to normalize it
			link.WithURL(u.String()),
		}

		// Secret required?
		if req.Secret != nil {
			// Derive secret hash from the secret value.
			sh, err := secretEncoder.Hash(*req.Secret)
			if err != nil {
				res.Error = &errorsv1.Error{
					ErrorMessage: "Unable to compute secret hash.",
					ErrorCode:    http.StatusInternalServerError,
				}
				return &res, err
			}

			// Add the secret hash.
			dopts = append(dopts, link.WithSecretHash(sh))
		}

		// Create a new Link.
		domainObject := link.New(dopts...)

		// Save to persistence
		if err := links.Save(ctx, domainObject); err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to create the shortened URL.",
				ErrorCode:    http.StatusInternalServerError,
			}
			return &res, err
		}

		// Prepare response
		res.Link = fromLink(domainObject)

		// Publish creation notification
		publisher.Publish(ctx, link.Created(res.Link))

		// No error
		return &res, nil
	}
}
