package serr

import errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"

type ServiceError interface {
	GetError() *errorsv1.Error
}
