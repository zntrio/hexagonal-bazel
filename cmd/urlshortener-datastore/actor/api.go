package actor

import (
	"context"
	"io"
)

type Actor interface {
	io.Closer

	Run(ctx context.Context) error
}
