package reactor

import "context"

// Handler describes the reactor message handler.
type Handler[REQ, RES any] func(ctx context.Context, req *REQ) (*RES, error)

// Decorator is used to decorate an handler.
type Decorator[REQ, RES any] func(ctx context.Context, req *REQ, res *RES) error
