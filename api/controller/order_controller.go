package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/dyaksa/warehouse/pkg/response/response_error"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderController struct {
	OrderUsecase domain.OrderUsecase
}

func (oc *OrderController) Checkout(c *gin.Context) {
	var body domain.CheckoutInput
	body.UserID = c.GetString("x-user-id")

	if err := c.ShouldBindJSON(&body); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("validation failed").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	result, err := oc.OrderUsecase.Checkout(c.Request.Context(), body)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("internal server error").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	response_success.JSON(c).Msg("success checkout order").Status("success").Data(result).Send(http.StatusOK)
}

// ConfirmPayment handles payment confirmation for an order
func (oc *OrderController) ConfirmPayment(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		response_error.JSON(c).Msg("invalid order ID format").Status("validation failed").Send(http.StatusBadRequest)
		return
	}

	err = oc.OrderUsecase.ConfirmPayment(c.Request.Context(), orderID)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("payment confirmation failed").Send(http.StatusBadRequest)
		return
	}

	response_success.JSON(c).Msg("payment confirmed successfully").Status("success").Send(http.StatusOK)
}

// CancelOrder handles order cancellation
func (oc *OrderController) CancelOrder(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		response_error.JSON(c).Msg("invalid order ID format").Status("validation failed").Send(http.StatusBadRequest)
		return
	}

	err = oc.OrderUsecase.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("order cancellation failed").Send(http.StatusBadRequest)
		return
	}

	response_success.JSON(c).Msg("order cancelled successfully").Status("success").Send(http.StatusOK)
}

// GetOrderDetails retrieves order details
func (oc *OrderController) GetOrderDetails(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		response_error.JSON(c).Msg("invalid order ID format").Status("validation failed").Send(http.StatusBadRequest)
		return
	}

	order, err := oc.OrderUsecase.GetOrderDetails(c.Request.Context(), orderID)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("order not found").Send(http.StatusNotFound)
		return
	}

	response_success.JSON(c).Msg("order details retrieved successfully").Status("success").Data(order).Send(http.StatusOK)
}

// GetUserOrders retrieves all orders for the authenticated user with pagination
func (oc *OrderController) GetUserOrders(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr := c.GetString("x-user-id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response_error.JSON(c).Msg("invalid user ID").Status("authentication error").Send(http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	var pagination paginator.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("validation failed").Send(http.StatusBadRequest)
		return
	}

	// Get user orders
	result, err := oc.OrderUsecase.GetUserOrders(c.Request.Context(), userID, pagination)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("failed to retrieve orders").Send(http.StatusInternalServerError)
		return
	}

	response_success.JSON(c).Msg("orders retrieved successfully").Status("success").Data(result).Send(http.StatusOK)
}
