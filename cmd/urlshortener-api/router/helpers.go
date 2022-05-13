package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	errorsv1 "zntr.io/hexagonal-bazel/api/system/errors/v1"
	"zntr.io/hexagonal-bazel/infrastructure/serr"
	"zntr.io/hexagonal-bazel/pkg/types"
)

func delegateTo[REQ any, RES serr.ServiceError](ctx context.Context, w http.ResponseWriter, req REQ, callFunc func(context.Context, REQ, ...grpc.CallOption) (RES, error)) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Call the handler
	res, err := callFunc(ctx, req)
	if err != nil {
		st := status.Convert(err)
		for _, detail := range st.Details() {
			switch t := detail.(type) {
			case *errorsv1.Error:
				// Set the HTTP code
				w.WriteHeader(int(t.StatusCode))

				// Send error as json
				if err := json.NewEncoder(w).Encode(t); err != nil {
					http.Error(w, "Unable to serialize service error", http.StatusInternalServerError)
					return
				}

				return
			default:
				continue
			}
		}
	}
	if !types.IsNil(res) {
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
}
