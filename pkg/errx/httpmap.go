package errx

import "net/http"

// HTTPStatus maps an error Code to an HTTP status code.
func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidArgument, CodeValidation:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists, CodeConflict:
		return http.StatusConflict
	case CodeUnauthenticated, CodeUnauthorized:
		return http.StatusUnauthorized
	case CodePermission:
		return http.StatusForbidden
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodePrecondition:
		return http.StatusPreconditionFailed
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// PublicMessage optionally maps internal error codes to safer public messages.
func PublicMessage(ae *AppError) string {
	if ae == nil {
		return ""
	}
	// In future we can provide custom messages per code.
	return ae.Message
}
