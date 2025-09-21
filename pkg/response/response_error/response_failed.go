package response_error

import "github.com/gin-gonic/gin"

type errorResponseContext struct {
	ctx *gin.Context
}

type errorResponse struct {
	errorResponseContext
	message string
	status  string
}

// Msg implements ErrorResponse.
func (e *errorResponse) Msg(msg string) *errorResponse {
	e.message = msg
	return e
}

// Send implements ErrorResponse.
func (e *errorResponse) Send(code int) {
	data := gin.H{
		"message": e.message,
		"status":  e.status,
	}

	e.ctx.JSON(code, data)
}

// Status implements ErrorResponse.
func (e *errorResponse) Status(status string) *errorResponse {
	e.status = status
	return e
}

type ErrorResponse interface {
	Msg(msg string) *errorResponse
	Status(status string) *errorResponse
	Send(code int)
}

func JSON(ctx *gin.Context) ErrorResponse {
	return &errorResponse{errorResponseContext: errorResponseContext{ctx: ctx}}
}
