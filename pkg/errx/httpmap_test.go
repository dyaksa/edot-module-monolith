package errx

import "testing"

func TestHTTPStatusMapping(t *testing.T) {
	cases := []struct {
		code     Code
		expected int
	}{
		{CodeValidation, 400},
		{CodeInvalidArgument, 400},
		{CodeNotFound, 404},
		{CodeAlreadyExists, 409},
		{CodeConflict, 409},
		{CodeUnauthenticated, 401},
		{CodePermission, 403},
		{CodeRateLimited, 429},
		{CodePrecondition, 412},
		{CodeUnavailable, 503},
		{CodeTimeout, 504},
		{CodeInternal, 500},
	}
	for _, c := range cases {
		if got := HTTPStatus(c.code); got != c.expected {
			t.Errorf("code %s expected %d got %d", c.code, c.expected, got)
		}
	}
}

func TestToAppError(t *testing.T) {
	orig := New(CodeNotFound, "resource missing")
	wrapped := ToAppError(orig)
	if wrapped != orig {
		t.Fatal("expected passthrough for AppError")
	}
}
