# Error Taxonomy & Central HTTP Mapping

This project now uses a structured error model (`pkg/errx`) plus a centralized HTTP translation middleware.

## Goals

- Consistent machine-readable error codes
- Single place to map codes -> HTTP status
- Separation of internal messages vs public payload (extensible via `PublicMessage`)
- Composable wrapping with operations and metadata

## Core Types

`errx.AppError` fields:

- `Code` (`errx.Code`): stable enum string (e.g. `NOT_FOUND`, `VALIDATION_ERROR`)
- `Message`: developer-oriented message (also sent publicly for now)
- `Op`: operation or logical function boundary (e.g. `UserUsecase.Login`)
- `Err`: wrapped underlying error (not serialized)
- `Meta`: optional contextual key/values (serialized)
- `Transient`: hint for retry semantics (not serialized)

## Defined Codes

```
INVALID_ARGUMENT
VALIDATION_ERROR
NOT_FOUND
ALREADY_EXISTS
UNAUTHENTICATED
UNAUTHORIZED
PERMISSION_DENIED
CONFLICT
RATE_LIMITED
PRECONDITION_FAILED
INTERNAL_ERROR
SERVICE_UNAVAILABLE
TIMEOUT
```

## HTTP Mapping

Handled in `errx.HTTPStatus`:

- 400: INVALID_ARGUMENT, VALIDATION_ERROR
- 401: UNAUTHENTICATED, UNAUTHORIZED
- 403: PERMISSION_DENIED
- 404: NOT_FOUND
- 409: ALREADY_EXISTS, CONFLICT
- 412: PRECONDITION_FAILED
- 429: RATE_LIMITED
- 500: INTERNAL_ERROR (default)
- 503: SERVICE_UNAVAILABLE
- 504: TIMEOUT

## Producing Errors

Use the ergonomic constructor:

```go
return errx.E(errx.CodeValidation, "email invalid", errx.Op("UserUsecase.Register"), underlyingErr)
```

Add metadata:

```go
return errx.E(errx.CodeConflict, "user already exists").WithMeta("email", input.Email)
```

## Controller Pattern

Instead of writing JSON directly for failures:

```go
if err != nil {
    c.Error(err) // controller returns; middleware serializes
    return
}
```

If you need to wrap on boundary:

```go
if err != nil {
    c.Error(errx.E(errx.CodeInternal, "failed to persist user", errx.Op("UserController.Create"), err))
    return
}
```

## Middleware Registration

Add `ErrorMiddleware` early in the Gin stack (after logger / rate limit but before auth-specific handlers is fine):

```go
r := gin.New()
r.Use(middleware.LoggerMiddleware(logger))
r.Use(middleware.ErrorMiddleware(logger))
```

## Response Shape

Errors now return:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request payload",
    "meta": { "field": "email" } // optional
  }
}
```

## Testing

`pkg/errx/httpmap_test.go` ensures code->status coverage. Add more tests around wrapping if needed.

## Migration Notes

- Existing `response_error` / `response_success` can coexist; gradually migrate other controllers.
- For legacy endpoints, you can wrap existing errors with `errx.ToAppError(err)` before calling `c.Error`.
- Eventually remove `response_error` after full migration.

## Future Enhancements

- Internationalization of public messages via `PublicMessage`
- Structured logging correlation IDs in `Meta`
- gRPC status mapping if needed
