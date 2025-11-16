package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

// CreateProductImage godoc
// @Summary      Add new image to product
// @Description  Upload a new image for a specific product
// @Tags         admin/products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true   "Bearer token"  default(Bearer <token>)
// @Param        id             path      int     true   "Product Id"
// @Param        image          formData  file    true   "Product image (JPEG/PNG, max 1MB)"
// @Param        isPrimary      formData  bool    false  "Set as primary image"  default(false)
// @Success      201  {object}  lib.ResponseSuccess{data=object{id=int,productImage=string}}  "Product image created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request"
// @Failure      404  {object}  lib.ResponseError  "Product not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /admin/products/{id}/images [post]
func CreateProductImage(ctx *gin.Context) {
	productId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	var bodyCreate models.ProductImageRequest
	err = ctx.ShouldBindWith(&bodyCreate, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
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

	// Get user id from token
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// Get and validate file
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Image is required",
			Error:   err.Error(),
		})
		return
	}

	if file.Size > 1<<20 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "File size must be less than 1MB",
		})
		return
	}

	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	if !allowedTypes[contentType] {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid file type. Only JPEG and PNG are allowed",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	if !allowedExt[ext] {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid file extension. Only JPG and PNG are allowed",
		})
		return
	}

	// Save file
	fileName := fmt.Sprintf("product_%d_%d%s", productId, time.Now().UnixNano(), ext)
	savedFilePath := "uploads/products/" + fileName

	os.MkdirAll("uploads/products", 0755)

	err = ctx.SaveUploadedFile(file, savedFilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to save uploaded file",
			Error:   err.Error(),
		})
		return
	}

	// Insert to database
	imageId, message, err := models.InsertProductImage(productId, savedFilePath, bodyCreate.IsPrimary, userId.(int))
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
		Data: gin.H{
			"id":           imageId,
			"productImage": savedFilePath,
			"isPrimary":    bodyCreate.IsPrimary,
		},
	})
}
