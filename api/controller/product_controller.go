package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
)

type ProductController struct {
	ProductUsecase domain.ProductUsecase
}

// Create creates a new product with initial stock in the specified warehouse
// @Summary Create a new product
// @Description Create a new product with SKU, name, and initial stock quantity in a warehouse
// @Tags Products
// @Accept json
// @Produce json
// @Param product body domain.CreateProductRequest true "Product creation data"
// @Success 201 {object} map[string]interface{} "Product created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload or validation failed"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /products [post]
func (pc *ProductController) Create(c *gin.Context) {
	var body domain.CreateProductRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid product payload", errx.Op("ProductController.Create"), err))
		return
	}

	if err := pc.ProductUsecase.Create(c.Request.Context(), body); err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("success create products").Status("success").Send(http.StatusCreated)
}

// RetrieveAll retrieves all products with pagination
// @Summary Get all products
// @Description Retrieve all products with pagination support and warehouse information
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "Products retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid pagination parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /products [get]
func (pc *ProductController) RetrieveAll(c *gin.Context) {
	var pagination paginator.PaginationRequest

	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid pagination query", errx.Op("ProductController.RetrieveAll"), err))
		return
	}

	result, err := pc.ProductUsecase.RetrieveAll(c.Request.Context(), pagination)

	if err != nil {
		c.Error(err)
		return
	}

	response_success.JSON(c).Msg("success retrieve products").Status("success").Data(result).Send(http.StatusOK)
}
