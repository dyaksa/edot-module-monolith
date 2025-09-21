package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/response/response_error"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
)

type ShopController struct {
	ShopUsecase domain.ShopUsecase
}

func (sc *ShopController) Create(c *gin.Context) {
	var body domain.CreateShopRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("error validation").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	if err := sc.ShopUsecase.Create(c.Request.Context(), body); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("internal server error").Send(http.StatusInternalServerError)
		c.Abort()
		return
	}

	response_success.JSON(c).Msg("success create shop").Status("success").Send(http.StatusCreated)
}

func (sc *ShopController) Retrieve(c *gin.Context) {
	var query domain.ShopQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("error validation").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	shop, err := sc.ShopUsecase.Retrieve(c.Request.Context(), query.ID)
	if err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("internal server error").Send(http.StatusInternalServerError)
		c.Abort()
		return
	}

	response_success.JSON(c).Msg("success retrieve shop").Data(shop).Send(http.StatusOK)
}

func (sc *ShopController) Update(c *gin.Context) {
	var body domain.UpdateShopRequest
	if err := c.ShouldBind(&body); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("error validation").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	if err := sc.ShopUsecase.Update(c.Request.Context(), body); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("internal server error").Send(http.StatusInternalServerError)
		c.Abort()
		return
	}

	response_success.JSON(c).Msg("success update shop").Status("success").Send(http.StatusOK)
}

func (sc *ShopController) Delete(c *gin.Context) {
	var query domain.ShopQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("error validation").Send(http.StatusBadRequest)
		c.Abort()
		return
	}

	if err := sc.ShopUsecase.Delete(c.Request.Context(), query.ID); err != nil {
		response_error.JSON(c).Msg(err.Error()).Status("internal server error").Send(http.StatusInternalServerError)
		c.Abort()
		return
	}

	response_success.JSON(c).Msg("success delete shop").Status("success").Send(http.StatusOK)
}
