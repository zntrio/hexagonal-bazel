package reactor

import (
	"context"
)

// Before executes the decorator before the handler
func Before[REQ, RES any](next Handler[REQ, RES], decorator Decorator[REQ, RES]) Handler[REQ, RES] {
	return func(ctx context.Context, req *REQ) (*RES, error) {
		res := new(RES)

		// Delegate to decorator
		if err := decorator(ctx, req, res); err != nil {
			return res, err
		}

		return next(ctx, req)
	}
}
