package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"zntr.io/hexagonal-bazel/infrastructure/serr"
	"zntr.io/hexagonal-bazel/pkg/types"
)

func delegateTo[REQ any, RES serr.ServiceError](ctx context.Context, w http.ResponseWriter, req REQ, callFunc func(context.Context, REQ, ...grpc.CallOption) (RES, error)) {
	// Call the handler
	res, _ := callFunc(ctx, req)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Public error
	switch {
	case types.IsNil(res):
		http.Error(w, "Unexpected nil response", http.StatusInternalServerError)
		return
	case res.GetError() != nil:
		// Set the HTTP code
		w.WriteHeader(int(res.GetError().StatusCode))

		// Send error as json
		if err := json.NewEncoder(w).Encode(res.GetError()); err != nil {
			http.Error(w, "Unable to serialize service error", http.StatusInternalServerError)
			return
		}
		return
	default:
		// This is fine
	}

	// Check if it's a protobuf message
	if msg, ok := any(res).(proto.Message); ok {
		payload, err := protojson.Marshal(msg)
		if err != nil {
			http.Error(w, "Unable to serialize response", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, string(payload))
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Unable to serialize response", http.StatusInternalServerError)
		return
	}
}
