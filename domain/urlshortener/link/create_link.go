package link

import (
	"context"
	"net/http"
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
	"zntr.io/hexagonal-bazel/pkg/reactor"
)

// CreateHandler handles the urlshortener.Create request.
func CreateHandler(links Repository, publisher eventbus.EventPublisher) reactor.Handler[urlshortener.CreateRequest, urlshortener.CreateResponse] {
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

		// Create a new Link.
		domainObject := New(
			// Generate a new identifier
			WithID(NewID()),
			// Use parsed URL to normalize it
			WithURL(u.String()),
		)

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
			Id:  string(domainObject.GetID()),
			Url: domainObject.GetURL(),
		}

		// Publish creation notification
		publisher.Publish(ctx, LinkCreated(res.Link))

		// No error
		return &res, nil
	}
}
