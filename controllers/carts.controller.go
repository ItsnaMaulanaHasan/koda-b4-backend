package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"net/http"
	"strconv"

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
	// get user id from token
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// get list carts
	carts, message, err := models.GetListCart(userId.(int))
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
		Data:    carts,
	})
}

// AddCart       godoc
// @Summary      Add new cart
// @Description  Add a new cart to list carts of user
// @Tags         carts
// @Accept       json
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
	err := ctx.ShouldBindJSON(&bodyAdd)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid JSON body",
			Error:   err.Error(),
		})
		return
	}

	// get user id from token
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}
	bodyAdd.UserId = userId.(int)

	// add data body to cart
	responseCart, message, err := models.AddToCart(bodyAdd)
	if err != nil {
		if message == "invalid amount, must be greater than 0" ||
			message == "amount exceeds available stock" {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: message,
			})
			return
		}

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

// DeleteCart    godoc
// @Summary      Delete cart
// @Description  Delete cart by Id
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "cart Id"
// @Success      200  {object}  lib.ResponseSuccess  "Cart deleted successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Cart not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while deleting cart data"
// @Router       /carts/{id} [delete]
func DeleteCart(ctx *gin.Context) {
	cartId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to convert type id from param",
			Error:   err.Error(),
		})
		return
	}

	// delete cart
	commandTag, err := models.DeleteCartById(cartId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while deleting cart",
			Error:   err.Error(),
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Cart not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Cart deleted successfully",
	})
}
