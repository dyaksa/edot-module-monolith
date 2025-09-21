package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/paginator"
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
		c.Error(errx.E(errx.CodeValidation, "invalid checkout payload", errx.Op("OrderController.Checkout"), err))
		return
	}

	result, err := oc.OrderUsecase.Checkout(c.Request.Context(), body)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("success checkout order").Status("success").Data(result).Send(http.StatusOK)
}

// ConfirmPayment handles payment confirmation for an order
func (oc *OrderController) ConfirmPayment(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.Error(err)
		return
	}

	err = oc.OrderUsecase.ConfirmPayment(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("payment confirmed successfully").Status("success").Send(http.StatusOK)
}

// CancelOrder handles order cancellation
func (oc *OrderController) CancelOrder(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.Error(err)
		return
	}

	err = oc.OrderUsecase.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("order cancelled successfully").Status("success").Send(http.StatusOK)
}

// GetOrderDetails retrieves order details
func (oc *OrderController) GetOrderDetails(c *gin.Context) {
	orderIDParam := c.Param("orderID")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.Error(err)
		return
	}

	order, err := oc.OrderUsecase.GetOrderDetails(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
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
		c.Error(errx.E(errx.CodeValidation, "invalid user ID", errx.Op("OrderController.GetUserOrders"), err))
		return
	}

	// Parse pagination parameters
	var pagination paginator.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid pagination query", errx.Op("OrderController.GetUserOrders"), err))
		return
	}

	// Get user orders
	result, err := oc.OrderUsecase.GetUserOrders(c.Request.Context(), userID, pagination)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("orders retrieved successfully").Status("success").Data(result).Send(http.StatusOK)
}
