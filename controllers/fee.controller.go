package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAllPaymentMethods godoc
// @Summary      Get all payment methods
// @Description  Retrieving all payment methods with admin fee
// @Tags         fees
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Success      200  {object}  lib.ResponseSuccess{data=[]models.PaymentMethod}  "Successfully retrieved payment methods"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /payment-methods [get]
func GetAllPaymentMethods(ctx *gin.Context) {
	paymentMethods, message, err := models.GetAllPaymentMethods()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data:    paymentMethods,
	})
}

// GetAllOrderMethods godoc
// @Summary      Get all order methods
// @Description  Retrieving all order methods with delivery fee
// @Tags         fees
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Success      200  {object}  lib.ResponseSuccess{data=[]models.OrderMethod}  "Successfully retrieved order methods"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /order-methods [get]
func GetAllOrderMethods(ctx *gin.Context) {
	orderMethods, message, err := models.GetAllOrderMethods()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data:    orderMethods,
	})
}
