package route

import (
	"time"

	"github.com/dyaksa/warehouse/api/controller"
	"github.com/dyaksa/warehouse/api/middleware"
	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/dyaksa/warehouse/repository"
	"github.com/dyaksa/warehouse/usecase"
	"github.com/gin-gonic/gin"
)

func NewOrderRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	jwtMiddleware := middleware.JwtAuthMiddleware(env.JwtSecret)
	orderRepository := repository.NewOrderRepository(db)
	idempotencyRepository := repository.NewIdempotencyRequestRepository(db)
	orderItemRepository := repository.NewOrderItemRepository(db)
	reservationRepository := repository.NewReservationRepository(db)
	movementRepository := repository.NewMovementRepository(db)
	productStockRepository := repository.NewProductStockRepository(db)
	pickWarehouseRepository := repository.NewWarehouseRepository(db)

	orderController := controller.OrderController{
		OrderUsecase: usecase.NewOrderUsecase(
			db.Database(),
			orderRepository,
			idempotencyRepository,
			orderItemRepository,
			reservationRepository,
			movementRepository,
			productStockRepository,
			pickWarehouseRepository,
		),
	}

	groupOrder := group.Group("/order", jwtMiddleware)
	groupOrder.POST("/checkout", orderController.Checkout)
	groupOrder.POST("/:orderID/confirm-payment", orderController.ConfirmPayment)
	groupOrder.POST("/:orderID/cancel", orderController.CancelOrder)
	groupOrder.GET("/:orderID", orderController.GetOrderDetails)
	groupOrder.GET("/list", orderController.GetUserOrders)
}
