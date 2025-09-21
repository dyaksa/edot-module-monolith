package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/api/worker"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
)

type StockReleaseController struct {
	stockReleaseWorker *worker.StockReleaseWorker
}

func NewStockReleaseController(stockReleaseWorker *worker.StockReleaseWorker) *StockReleaseController {
	return &StockReleaseController{
		stockReleaseWorker: stockReleaseWorker,
	}
}

// ManualRelease manually triggers the stock release process
func (src *StockReleaseController) ManualRelease(c *gin.Context) {
	err := src.stockReleaseWorker.ProcessNow(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process stock release",
			"message": err.Error(),
		})
		return
	}

	response_success.JSON(c).
		Status("success").
		Msg("Stock release processing triggered successfully").
		Send(http.StatusOK)
}

// Status returns the status of the stock release worker
func (src *StockReleaseController) Status(c *gin.Context) {
	response_success.JSON(c).
		Status("success").
		Msg("Stock release worker is running").
		Data(gin.H{
			"worker_status": "active",
		}).
		Send(http.StatusOK)
}
