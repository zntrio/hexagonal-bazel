package serr

import "net/http"

// ServerError returns a compliant `server_error` error.
func ServerError(err error) Builder {
	return &defaultErrorBuilder{
		opts: &Options{
			statusCode:       http.StatusInternalServerError,
			errorCode:        "server_error",
			errorDescription: "The server encountered an unexpected condition that prevented it from fulfilling the request.",
			internalError:    err,
		},
	}
}

// InvalidRequest returns a compliant `invalid_request` error.
func InvalidRequest() Builder {
	return &defaultErrorBuilder{
		opts: &Options{
			statusCode:       http.StatusBadRequest,
			errorCode:        "invalid_request",
			errorDescription: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed.",
		},
	}
}

// ResourceNotFound returns a compliant `resource_not_found` error.
func ResourceNotFound() Builder {
	return &defaultErrorBuilder{
		opts: &Options{
			statusCode:       http.StatusNotFound,
			errorCode:        "resource_not_found",
			errorDescription: "The requested resource could not be resolved, is expired or not found.",
		},
	}
}

// AccessDenied returns a compliant `access_denied` error.
func AccessDenied() Builder {
	return &defaultErrorBuilder{
		opts: &Options{
			statusCode:       http.StatusForbidden,
			errorCode:        "access_denied",
			errorDescription: "The authorization request was denied.",
		},
	}
}
