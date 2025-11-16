package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListProductImages godoc
// @Summary      Get all images of a product
// @Description  Retrieving all images of a specific product
// @Tags         admin/products
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Product Id"
// @Success      200  {object}  lib.ResponseSuccess{data=[]models.ProductImage}  "Successfully retrieved product images"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Product not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /admin/products/{id}/images [get]
func ListProductImages(ctx *gin.Context) {
	productId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	// Check if product exists
	exists, err := models.CheckProductExists(productId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking product existence",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Product not found",
		})
		return
	}

	// Get product images
	images, message, err := models.GetProductImages(productId)
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
		Data:    images,
	})
}

// DetailProductImage godoc
// @Summary      Get product image by Id
// @Description  Retrieving specific product image based on image Id
// @Tags         admin/products
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Product Id"
// @Param        imageId        path    int     true  "Image Id"
// @Success      200  {object}  lib.ResponseSuccess{data=models.ProductImage}  "Successfully retrieved product image"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Product image not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /admin/products/{id}/images/{imageId} [get]
func DetailProductImage(ctx *gin.Context) {
	imageId, err := strconv.Atoi(ctx.Param("imageId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid image Id format",
			Error:   err.Error(),
		})
		return
	}

	// Get product image
	image, message, err := models.GetProductImageById(imageId)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Product image not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data:    image,
	})
}
