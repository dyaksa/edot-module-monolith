package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WarehouseController struct {
	WarehouseUsecase domain.WarehouseUsecase
}

// Create creates a new warehouse for a shop
// @Summary Create a new warehouse
// @Description Create a new warehouse associated with a shop
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param warehouse body domain.WarehouseCreateRequest true "Warehouse creation data"
// @Success 201 {object} map[string]interface{} "Warehouse created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload or validation failed"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses [post]
func (w *WarehouseController) Create(c *gin.Context) {
	var body domain.WarehouseCreateRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse payload", errx.Op("WarehouseController.Create"), err))
		return
	}

	if err := w.WarehouseUsecase.Create(c.Request.Context(), body); err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("success create warehouse").Status("success").Send(http.StatusCreated)
}

// Retrieve gets a warehouse by ID
// @Summary Get warehouse by ID
// @Description Retrieve a specific warehouse by its UUID
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} "Warehouse retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid warehouse ID format"
// @Failure 404 {object} map[string]interface{} "Warehouse not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses/{id} [get]
func (w *WarehouseController) Retrieve(c *gin.Context) {
	warehouseIDStr := c.Param("id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse ID", errx.Op("WarehouseController.Retrieve"), err))
		return
	}

	warehouse, err := w.WarehouseUsecase.Retrieve(c.Request.Context(), warehouseID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Data(warehouse).Msg("warehouse retrieved successfully").Status("success").Send(http.StatusOK)
}

// Update updates a warehouse by ID
// @Summary Update warehouse
// @Description Update an existing warehouse by its UUID
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID (UUID)" format(uuid)
// @Param warehouse body domain.WarehouseCreateRequest true "Updated warehouse data"
// @Success 200 {object} map[string]interface{} "Warehouse updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid warehouse ID or request payload"
// @Failure 404 {object} map[string]interface{} "Warehouse not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses/{id} [put]
func (w *WarehouseController) Update(c *gin.Context) {
	warehouseIDStr := c.Param("id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse ID", errx.Op("WarehouseController.Update"), err))
		return
	}

	var body domain.WarehouseCreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse payload", errx.Op("WarehouseController.Update"), err))
		return
	}

	if err := w.WarehouseUsecase.Update(c.Request.Context(), warehouseID, body); err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("warehouse updated successfully").Status("success").Send(http.StatusOK)
}

// Delete deletes a warehouse by ID
// @Summary Delete warehouse
// @Description Delete an existing warehouse by its UUID
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} "Warehouse deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid warehouse ID format"
// @Failure 404 {object} map[string]interface{} "Warehouse not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses/{id} [delete]
func (w *WarehouseController) Delete(c *gin.Context) {
	warehouseIDStr := c.Param("id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse ID", errx.Op("WarehouseController.Delete"), err))
		return
	}

	if err := w.WarehouseUsecase.Delete(c.Request.Context(), warehouseID); err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("warehouse deleted successfully").Status("success").Send(http.StatusOK)
}

// SetActive updates warehouse active status
// @Summary Set warehouse active status
// @Description Enable or disable a warehouse by updating its active status
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID (UUID)" format(uuid)
// @Param status body object{is_active=bool} true "Active status data"
// @Success 200 {object} map[string]interface{} "Warehouse status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid warehouse ID or request payload"
// @Failure 404 {object} map[string]interface{} "Warehouse not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /warehouses/{id}/status [patch]
func (w *WarehouseController) SetActive(c *gin.Context) {
	warehouseIDStr := c.Param("id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid warehouse ID", errx.Op("WarehouseController.SetActive"), err))
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid status payload", errx.Op("WarehouseController.SetActive"), err))
		return
	}

	if err := w.WarehouseUsecase.SetActive(c.Request.Context(), warehouseID, req.IsActive); err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("warehouse status updated successfully").Status("success").Send(http.StatusOK)
}

// GetByShop gets warehouses by shop ID
// @Summary Get warehouses by shop ID
// @Description Retrieve all warehouses associated with a specific shop
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param shop_id path string true "Shop ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} "Warehouses retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid shop ID format"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /shops/{shop_id}/warehouses [get]
func (w *WarehouseController) GetByShop(c *gin.Context) {
	shopIDStr := c.Param("shop_id")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid shop ID", errx.Op("WarehouseController.GetByShop"), err))
		return
	}

	warehouses, err := w.WarehouseUsecase.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Data(warehouses).Msg("warehouses retrieved successfully").Status("success").Send(http.StatusOK)
}
