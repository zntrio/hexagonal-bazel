package link

import (
	"context"
	"net/http"
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/infrastructure/generator"
	"zntr.io/hexagonal-bazel/infrastructure/security/password"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
	"zntr.io/hexagonal-bazel/pkg/reactor"
	"zntr.io/hexagonal-bazel/pkg/types"
)

// CreateHandler handles the urlshortener.Create request.
func CreateHandler(links Repository, publisher eventbus.EventPublisher, codeGenerator generator.Generator[string], secretEncoder password.Hasher) reactor.Handler[urlshortener.CreateRequest, urlshortener.CreateResponse] {
	return func(ctx context.Context, req *urlshortener.CreateRequest) (*urlshortener.CreateResponse, error) {
		var res urlshortener.CreateResponse

		// Check arguments
		if req == nil {
			return &res, nil
		}

		// Validate request
		if err := validation.ValidateStruct(req,
			// URL is mandatoray and must contain a valid URL syntax.
			validation.Field(&req.Url, validation.Required, is.URL),
		); err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to validate request execution conditions.",
				ErrorCode:    http.StatusPreconditionFailed,
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
			return &res, nil
		}

		// Skip operation if in validation mode
		if req.ValidateOnly {
			return &res, nil
		}

		// Create a public identifier
		code, err := codeGenerator.Generate()
		if err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to generate shortened identifier.",
				ErrorCode:    http.StatusInternalServerError,
			}
			return &res, nil
		}

		// Set required properties
		dopts := []DomainOption{
			// Generate a new identifier
			WithID(ID(code)),
			// Use parsed URL to normalize it
			WithURL(u.String()),
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
				return &res, nil
			}

			// Add the secret hash.
			dopts = append(dopts, WithSecretHash(sh))
		}

		// Create a new Link.
		domainObject := New(dopts...)

		// Save to persistence
		if err := links.Save(ctx, domainObject); err != nil {
			res.Error = &errorsv1.Error{
				ErrorMessage: "Unable to create the shortened URL.",
				ErrorCode:    http.StatusInternalServerError,
			}
			return &res, err
		}

		// Prepare response
		res.Link = &urlshortener.Link{
			Id:             string(domainObject.GetID()),
			Url:            types.AsRef(domainObject.GetURL()),
			SecretRequired: domainObject.GetSecretHash() != "",
		}

		// Publish creation notification
		publisher.Publish(ctx, LinkCreated(res.Link))

		// No error
		return &res, nil
	}
}
