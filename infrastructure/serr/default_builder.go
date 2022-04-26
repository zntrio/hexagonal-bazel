package serr

import (
	"errors"
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	"zntr.io/hexagonal-bazel/pkg/types"
)

type defaultErrorBuilder struct {
	opts *Options
}

func (eb *defaultErrorBuilder) Build(opts ...Option) *errorsv1.Error {
	if eb.opts == nil {
		eb.opts = &Options{}
	}

	// Apply options
	for _, o := range opts {
		o(eb.opts)
	}

	// Create error object
	err := &errorsv1.Error{
		StatusCode:       int32(eb.opts.statusCode),
		ErrorCode:        eb.opts.errorCode,
		ErrorDescription: eb.opts.errorDescription,
	}
	if eb.opts.errorURI != "" {
		err.ErrorUri = eb.opts.errorURI
	}
	if len(eb.opts.fields) > 0 {
		err.Fields = eb.opts.fields
	}

	// Return error instance
	return err
}

// -----------------------------------------------------------------------------

func fieldsFromError(err error) []string {
	var errs validation.Errors
	// Check if it's a validation error
	if errors.As(err, &errs) {
		var fields types.StringArray
		for k := range errs {
			fields.AddIfNotContains(k)
		}
		sort.Strings(fields)
		return fields
	}

	// No fields
	return nil
}
