package reactor

import "context"

// Handler describes the reactor message handler.
type Handler[REQ, RES any] func(ctx context.Context, req *REQ) (*RES, error)
