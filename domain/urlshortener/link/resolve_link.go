package link

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	urlshortener "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"zntr.io/hexagonal-bazel/pkg/reactor"
)

// ResolveHandler handles the urlshortener.Resolve request.
func ResolveHandler(links Resolver) reactor.Handler[urlshortener.ResolveRequest, urlshortener.ResolveResponse] {
	return func(ctx context.Context, req *urlshortener.ResolveRequest) (*urlshortener.ResolveResponse, error) {
		var res urlshortener.ResolveResponse

		// Check arguments
		if req == nil {
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
		m, err := links.GetByID(ctx, ID(req.Id))
		switch {
		case err == nil:
			// Skip processing
		case errors.Is(err, ErrLinkNotFound):
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

		// Prepare response
		res.Link = &urlshortener.Link{
			Id:  string(m.GetID()),
			Url: m.GetURL(),
		}

		// No error
		return &res, nil
	}
}
