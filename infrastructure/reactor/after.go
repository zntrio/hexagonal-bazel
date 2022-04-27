package reactor

import (
	"context"
)

// After executes the decorator after the handler.
func After[REQ, RES any](next Handler[REQ, RES], decorator Decorator[REQ, RES]) Handler[REQ, RES] {
	return func(ctx context.Context, req *REQ) (*RES, error) {
		res, err := next(ctx, req)
		if err != nil {
			return res, err
		}

		// Delegate to decorator
		if err := decorator(ctx, req, res); err != nil {
			return res, err
		}

		return res, err
	}
}
