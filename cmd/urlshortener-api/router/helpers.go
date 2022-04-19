package router

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"zntr.io/hexagonal-bazel/pkg/serr"
	"zntr.io/hexagonal-bazel/pkg/types"
)

func delegateTo[REQ any, RES serr.ServiceError](ctx context.Context, w http.ResponseWriter, req REQ, callFunc func(context.Context, REQ, ...grpc.CallOption) (RES, error)) {
	// Call the handler
	res, err := callFunc(ctx, req)
	switch {
	case err != nil:
		http.Error(w, "Unexpected error occured", http.StatusInternalServerError)
		return
	case types.IsNil(res):
		http.Error(w, "Unexpected nil response", http.StatusInternalServerError)
		return
	case res.GetError() != nil:
		// Set the HTTP code
		w.WriteHeader(int(res.GetError().ErrorCode))

		// Send error as json
		if err := json.NewEncoder(w).Encode(res.GetError()); err != nil {
			http.Error(w, "Unable to serialize service error", http.StatusInternalServerError)
			return
		}
		return
	default:
		// This is fine
	}

	// Send response
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Unable to serialize response", http.StatusInternalServerError)
		return
	}
}
