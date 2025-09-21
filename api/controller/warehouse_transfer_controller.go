package controller

import (
	"net/http"
	"strconv"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WarehouseTransferController struct {
	TransferUsecase domain.WarehouseTransferUsecase
}

// CreateTransfer creates a new warehouse transfer
// @Summary Create a new warehouse transfer
// @Description Create a transfer request to move products between warehouses with specified items and quantities
// @Tags Warehouse Transfers
// @Accept json
// @Produce json
// @Param transfer body domain.CreateTransferRequest true "Transfer creation data with items"
// @Success 201 {object} map[string]interface{} "Transfer created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload or validation failed"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /transfers [post]
func (wtc *WarehouseTransferController) CreateTransfer(c *gin.Context) {
	var req domain.CreateTransferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid transfer payload", errx.Op("WarehouseTransferController.CreateTransfer"), err))
		return
	}

	transfer, err := wtc.TransferUsecase.CreateTransfer(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Data(transfer).Msg("transfer created successfully").Status("success").Send(http.StatusCreated)
}

// GetTransfer retrieves a warehouse transfer by ID
// @Summary Get warehouse transfer by ID
// @Description Retrieve a specific warehouse transfer with all its details and items
// @Tags Warehouse Transfers
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} "Transfer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid transfer ID format"
// @Failure 404 {object} map[string]interface{} "Transfer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /transfers/{id} [get]
func (wtc *WarehouseTransferController) GetTransfer(c *gin.Context) {
	transferIDStr := c.Param("id")
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid transfer ID", errx.Op("WarehouseTransferController.GetTransfer"), err))
		return
	}

	transfer, err := wtc.TransferUsecase.GetTransfer(c.Request.Context(), transferID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Data(transfer).Msg("transfer retrieved successfully").Status("success").Send(http.StatusOK)
}

// UpdateTransferStatus updates the status of a warehouse transfer
// @Summary Update transfer status
// @Description Update the status of a warehouse transfer (REQUESTED, APPROVED, IN_TRANSIT, COMPLETED, CANCELLED)
// @Tags Warehouse Transfers
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID (UUID)" format(uuid)
// @Param status body domain.UpdateTransferStatusRequest true "New transfer status"
// @Success 200 {object} map[string]interface{} "Transfer status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid transfer ID or status"
// @Failure 404 {object} map[string]interface{} "Transfer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /transfers/{id}/status [patch]
func (wtc *WarehouseTransferController) UpdateTransferStatus(c *gin.Context) {
	transferIDStr := c.Param("id")
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid transfer ID", errx.Op("WarehouseTransferController.UpdateTransferStatus"), err))
		c.Abort()
		return
	}

	var req domain.UpdateTransferStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid status payload", errx.Op("WarehouseTransferController.UpdateTransferStatus"), err))
		return
	}

	err = wtc.TransferUsecase.UpdateTransferStatus(c.Request.Context(), transferID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("transfer status updated successfully").Status("success").Send(http.StatusOK)
}

// GetTransfersByWarehouse retrieves transfers associated with a warehouse
// @Summary Get transfers by warehouse
// @Description Retrieve all transfers where the warehouse is either source or destination with pagination
// @Tags Warehouse Transfers
// @Accept json
// @Produce json
// @Param warehouse_id path string true "Warehouse ID (UUID)" format(uuid)
// @Param limit query int false "Number of items per page" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "Transfers retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid warehouse ID or pagination parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses/{warehouse_id}/transfers [get]
func (wtc *WarehouseTransferController) GetTransfersByWarehouse(c *gin.Context) {
	warehouseIDStr := c.Param("warehouse_id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse ID", errx.Op("WarehouseTransferController.GetTransfersByWarehouse"), err))
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	transfers, total, err := wtc.TransferUsecase.GetTransfersByWarehouse(c.Request.Context(), warehouseID, limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	result := map[string]interface{}{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}

	response_success.JSON(c).Data(result).Msg("transfers retrieved successfully").Status("success").Send(http.StatusOK)
}

// ExecuteTransfer executes a warehouse transfer
// @Summary Execute warehouse transfer
// @Description Execute an approved transfer to actually move stock between warehouses
// @Tags Warehouse Transfers
// @Accept json
// @Produce json
// @Param id path string true "Transfer ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} "Transfer executed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid transfer ID or transfer not in executable state"
// @Failure 404 {object} map[string]interface{} "Transfer not found"
// @Failure 500 {object} map[string]interface{} "Internal server error or execution failed"
// @Security BearerAuth
// @Router /transfers/{id}/execute [post]
func (wtc *WarehouseTransferController) ExecuteTransfer(c *gin.Context) {
	transferIDStr := c.Param("id")
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid transfer ID", errx.Op("WarehouseTransferController.ExecuteTransfer"), err))
		return
	}

	err = wtc.TransferUsecase.ExecuteTransfer(c.Request.Context(), transferID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("transfer executed successfully").Status("success").Send(http.StatusOK)
}
