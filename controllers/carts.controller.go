package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			c.id, 
			c.user_id, 
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS product_images, 
			p.name AS product_name, 
			s.name AS size_name, 
			v.name AS variant_name, 
			c.amount, 
			c.subtotal
			FROM carts c
		LEFT JOIN products p ON p.id = c.product_id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		LEFT JOIN sizes s  ON s.id = c.size_id
		LEFT JOIN variants v ON v.id = c.variant_id
		WHERE c.user_id = $1
		GROUP BY c.id, p.name, s.name, v.name
		ORDER BY c.updated_at DESC`, userIdFromToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch list carts from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	carts, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Cart])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process carts data from database",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get list carts",
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
		Message: "Cart added successfully",
		Data:    responseCart,
	})
}

func DeleteCart(ctx *gin.Context) {

}
