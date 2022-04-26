package serr

import (
	"fmt"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	"zntr.io/hexagonal-bazel/pkg/types"
)

type ServiceError interface {
	GetError() *errorsv1.Error
}

type Option func(opts *Options)

type Options struct {
	statusCode       int
	errorCode        string
	errorDescription string
	errorURI         string
	internalError    error
	fields           []string
}

// ErrorBuilder describes error builder contract.
type Builder interface {
	Build(opts ...Option) *errorsv1.Error
}

// ----------------------------------------------------------------------------

func Description(value string) Option {
	return func(opts *Options) {
		opts.errorDescription = value
	}
}

func Descriptionf(value string, args ...any) Option {
	return func(opts *Options) {
		opts.errorDescription = fmt.Sprintf(value, args...)
	}
}

func ErrorCode(value string) Option {
	return func(opts *Options) {
		opts.errorCode = value
	}
}

func ErrorURI(value string) Option {
	return func(opts *Options) {
		opts.errorURI = value
	}
}

func Fields(values ...string) Option {
	return func(opts *Options) {
		fields := types.StringArray(opts.fields)
		fields.AddIfNotContains(values...)
		opts.fields = fields.Unique()
	}
}

func InternalErr(err error) Option {
	return func(opts *Options) {
		opts.internalError = err
		opts.fields = fieldsFromError(err)
	}
}

func StatusCode(value int) Option {
	return func(opts *Options) {
		opts.statusCode = value
	}
}
