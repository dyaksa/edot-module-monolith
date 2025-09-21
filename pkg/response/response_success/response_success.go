package response_success

import "github.com/gin-gonic/gin"

type successResponseContext struct {
	ctx *gin.Context
}

type successResponse struct {
	successResponseContext
	message string
	status  string
	data    any
}

// Msg implements ErrorResponse.
func (e *successResponse) Msg(msg string) *successResponse {
	e.message = msg
	return e
}

// Send implements ErrorResponse.
func (e *successResponse) Send(code int) {
	data := gin.H{
		"data":    e.data,
		"message": e.message,
		"status":  e.status,
	}

	e.ctx.JSON(code, data)
}

// Status implements ErrorResponse.
func (e *successResponse) Status(status string) *successResponse {
	e.status = status
	return e
}

type SuccessResponse interface {
	Msg(msg string) *successResponse
	Status(status string) *successResponse
	Data(v any) *successResponse
	Send(code int)
}

func (e *successResponse) Data(v any) *successResponse {
	e.data = v
	return e
}

func JSON(ctx *gin.Context) SuccessResponse {
	return &successResponse{successResponseContext: successResponseContext{ctx: ctx}}
}
