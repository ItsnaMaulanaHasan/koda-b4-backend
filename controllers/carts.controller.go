package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListCarts     godoc
// @Summary      Get list carts
// @Description  Retrieving list cart by user id
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true   "Bearer token"    default(Bearer <token>)
// @Success      200  {object}  lib.ResponseSuccess{data=[]models.Cart} "Successfully retrieved carts list"
// @Failure      401  {object}  lib.ResponseError  "User unauthorized"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while fetching or processing carts data"
// @Router       /carts [get]
func ListCarts(ctx *gin.Context) {
	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	carts, message, err := models.GetListCart(userIdFromToken.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    carts,
	})
}

// AddCart       godoc
// @Summary      Add new cart
// @Description  Add a new cart to list carts of user
// @Tags         carts
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        dataCart       body    models.CartRequest  true  "Data request add cart"
// @Success      201            {object}  lib.ResponseSuccess{data=models.Cart}  "Cart added successfully"
// @Failure      400            {object}  lib.ResponseError  "Invalid request body"
// @Failure      401            {object}  lib.ResponseError  "User unauthorized"
// @Failure      500            {object}  lib.ResponseError  "Internal server error while adding, updating, or get data from cart"
// @Router       /carts [post]
func AddCart(ctx *gin.Context) {
	var bodyAdd models.CartRequest
	err := ctx.ShouldBindBodyWithJSON(&bodyAdd)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}
	bodyAdd.UserId = userIdFromToken.(int)

	responseCart, message, err := models.AddToCart(bodyAdd)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data:    responseCart,
	})
}

func DeleteCart(ctx *gin.Context) {

}
