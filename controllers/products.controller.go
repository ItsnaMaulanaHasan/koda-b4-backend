package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ListProductsAdmin  godoc
// @Summary           Get list products for admin
// @Description       Retrieving list products with pagination support and search
// @Tags              admin/products
// @Produce           json
// @Security          BearerAuth
// @Param             Authorization  header    string  true   "Bearer token" default(Bearer <token>)
// @Param             page   		 query     int     false  "Page number"  default(1)  minimum(1)
// @Param             limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(50)
// @Param             search         query     string  false  "Search value"
// @Success           200            {object}  object{success=bool,message=string,data=[]models.AdminProductResponse,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved product list"
// @Failure           400            {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure           500            {object}  lib.ResponseError  "Internal server error while fetching or processing product data."
// @Router            /admin/products [get]
func ListProductsAdmin(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")

	if page < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'page' must be greater than 0",
		})
		return
	}

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' must be greater than 0",
		})
		return
	}

	if limit > 100 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' cannot exceed 100",
		})
		return
	}

	var totalData int
	var err error
	rdb := lib.Redis()

	totalCacheKey := fmt.Sprintf("products:total:search:%s", search)

	// redis for total products
	cacheTotalDataProducts, _ := rdb.Get(context.Background(), totalCacheKey).Result()
	if cacheTotalDataProducts == "" {
		totalData, err = models.TotalDataProducts(search)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}
		err = rdb.Set(context.Background(), totalCacheKey, totalData, 15*time.Minute).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set total products to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cacheTotalDataProducts), &totalData)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal total products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	listCacheKey := fmt.Sprintf("products:list:page:%d:limit:%d:search:%s", page, limit, search)

	// redis for list products
	var products []models.AdminProductResponse
	cacheListAllProducts, _ := rdb.Get(context.Background(), listCacheKey).Result()
	if cacheListAllProducts == "" {
		products, err = models.GetListProductsAdmin(search, page, limit)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}

		productsStr, err := json.Marshal(products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization list all products",
				Error:   err.Error(),
			})
			return
		}

		err = rdb.Set(context.Background(), listCacheKey, productsStr, 15*time.Minute).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set list all products to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cacheListAllProducts), &products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal list all products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	// hateoas
	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/admin/products", scheme, host)

	var next any
	var prev any

	if totalData == 0 {
		page = 0
		next = nil
		prev = nil
	} else if page == 1 && totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	} else if page == totalPage && totalPage > 1 {
		next = nil
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else if totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else {
		next = nil
		prev = nil
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get all product",
		"data":    products,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
			"next":        next,
			"prev":        prev,
		},
	})
}

// DetailProductAdmin   godoc
// @Summary        		Get product by Id
// @Description    		Retrieving product data based on Id for admin
// @Tags           		admin/products
// @Accept 		   		x-www-form-urlencoded
// @Produce        		json
// @Security       		BearerAuth
// @Param          		Authorization    header  string  true  "Bearer token"  default(Bearer <token>)
// @Param          		id   			path    int     true  "product Id"
// @Success        		200  {object}  lib.ResponseSuccess{data=models.AdminProductResponse}  "Successfully retrieved product"
// @Failure        		400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure        		404  {object}  lib.ResponseError  "Product not found"
// @Failure        		500  {object}  lib.ResponseError  "Internal server error while fetching products from database"
// @Router         		/admin/products/{id} [get]
func DetailProductAdmin(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	rdb := lib.Redis()

	// redis for detail product
	var product models.AdminProductResponse
	cache, _ := rdb.Get(context.Background(), ctx.Request.RequestURI).Result()
	if cache == "" {
		product, message, err := models.GetDetailProductAdmin(id)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if message == "Product not found" {
				statusCode = http.StatusNotFound
			}
			ctx.JSON(statusCode, lib.ResponseError{
				Success: false,
				Message: message,
				Error:   err.Error(),
			})
			return
		}

		productStr, err := json.Marshal(product)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization detail product",
				Error:   err.Error(),
			})
			return
		}

		err = rdb.Set(context.Background(), ctx.Request.RequestURI, productStr, 60*time.Second).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set data product to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cache), &product)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal data products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get product",
		Data:    product,
	})
}

// CreateProduct godoc
// @Summary      Create new product
// @Description  Create a new product with images, sizes, and categories
// @Tags         admin/products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization      header    string    true   "Bearer token"  default(Bearer <token>)
// @Param        name               formData  string    true   "Product name"
// @Param        description        formData  string    true   "Product description"
// @Param        price              formData  number    true   "Product price"
// @Param        discountPercent    formData  number    true   "Discount percentage" default(0.00)
// @Param        rating			    formData  number    true   "Product rating" default(5)
// @Param        stock              formData  int       true   "Product stock"
// @Param        isFlashSale        formData  bool      true   "Is flash sale"  default(false)
// @Param        isActive           formData  bool      true   "Is active"  default(true)
// @Param        isFavourite        formData  bool      true   "Is active"  default(false)
// @Param        image1             formData  file      true   "Product image 1 (JPEG/PNG, max 1MB)"
// @Param        image2             formData  file      true   "Product image 2 (JPEG/PNG, max 1MB)"
// @Param        image3             formData  file      true   "Product image 3 (JPEG/PNG, max 1MB)"
// @Param        image4             formData  file      true   "Product image 4 (JPEG/PNG, max 1MB)"
// @Param        sizeProducts       formData  string    true   "Size Id (comma-separated, e.g., 1,2,3)"
// @Param        productCategories  formData  string    true   "Category Id (comma-separated, e.g., 1,2,3)"
// @Param        productVariants  	formData  string    true   "Variant Id (comma-separated, e.g., 1,2,3)"
// @Success      201  {object}  lib.ResponseSuccess{data=models.AdminProductResponse}  "Product created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      409  {object}  lib.ResponseError  "Product name already exists"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /admin/products [post]
func CreateProduct(ctx *gin.Context) {
	var bodyCreate models.ProductRequest
	err := ctx.ShouldBindWith(&bodyCreate, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	// check product with name from body is already exist
	exists, err := models.CheckProductName(bodyCreate.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking product name uniqueness",
			Error:   err.Error(),
		})
		return
	}
	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Product name already exists",
		})
		return
	}

	// get user id from token
	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// Start transaction
	tx, err := config.DB.Begin(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to start database transaction",
			Error:   err.Error(),
		})
		return
	}
	defer tx.Rollback(context.Background())

	// Insert product
	err = models.InsertDataProduct(tx, &bodyCreate, userIdFromToken.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new product",
			Error:   err.Error(),
		})
		return
	}

	// Process and save images
	fileImages := map[string]*multipart.FileHeader{
		"image1": bodyCreate.Image1,
		"image2": bodyCreate.Image2,
		"image3": bodyCreate.Image3,
		"image4": bodyCreate.Image4,
	}

	var savedImagePaths []string

	for _, file := range fileImages {
		if file == nil {
			continue
		}

		if file.Size > 1<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "Each file size must be less than 1MB",
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

		fileName := fmt.Sprintf("product_%d_%d%s", bodyCreate.Id, time.Now().UnixNano(), ext)
		savedFilePath := "uploads/products/" + fileName

		os.MkdirAll("uploads/products", 0755)

		err := ctx.SaveUploadedFile(file, savedFilePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to save uploaded file",
				Error:   err.Error(),
			})
			return
		}

		savedImagePaths = append(savedImagePaths, savedFilePath)
	}

	// Insert images
	if len(savedImagePaths) > 0 {
		err = models.InsertProductImages(tx, bodyCreate.Id, savedImagePaths, userIdFromToken.(int))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while inserting product images",
				Error:   err.Error(),
			})
			return
		}
	}
	bodyCreate.ProductImages = savedImagePaths

	// Insert sizes
	if strings.TrimSpace(bodyCreate.SizeProducts) != "" {
		sizeProducts := strings.Split(bodyCreate.SizeProducts, ",")
		var sizeIds []int
		for _, sizeIdStr := range sizeProducts {
			sizeIdStr = strings.TrimSpace(sizeIdStr)
			if sizeIdStr == "" {
				continue
			}
			sizeId, err := strconv.Atoi(sizeIdStr)
			if err != nil {
				continue
			}
			sizeIds = append(sizeIds, sizeId)
		}

		if len(sizeIds) > 0 {
			err = models.InsertProductSizes(tx, bodyCreate.Id, sizeIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting sizes product",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Insert categories
	if strings.TrimSpace(bodyCreate.ProductCategories) != "" {
		productCategories := strings.Split(bodyCreate.ProductCategories, ",")
		var categoryIds []int
		for _, categoryIdStr := range productCategories {
			categoryIdStr = strings.TrimSpace(categoryIdStr)
			if categoryIdStr == "" {
				continue
			}
			categoryId, err := strconv.Atoi(categoryIdStr)
			if err != nil {
				continue
			}
			categoryIds = append(categoryIds, categoryId)
		}

		if len(categoryIds) > 0 {
			err = models.InsertProductCategories(tx, bodyCreate.Id, categoryIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product categories",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Insert variants
	if strings.TrimSpace(bodyCreate.ProductVariants) != "" {
		productVariants := strings.Split(bodyCreate.ProductVariants, ",")
		var variantIds []int
		for _, variantIdStr := range productVariants {
			variantIdStr = strings.TrimSpace(variantIdStr)
			if variantIdStr == "" {
				continue
			}
			variantId, err := strconv.Atoi(variantIdStr)
			if err != nil {
				continue
			}
			variantIds = append(variantIds, variantId)
		}

		if len(variantIds) > 0 {
			err = models.InsertProductVariants(tx, bodyCreate.Id, variantIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product variants",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to commit transaction",
			Error:   err.Error(),
		})
		return
	}

	if err := models.InvalidateProductCache(context.Background()); err != nil {
		fmt.Printf("Warning: Failed to invalidate cache: %v\n", err)
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: "Product created successfully",
		Data:    bodyCreate,
	})
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Updating product data based on Id
// @Tags         admin/products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization      header    string    true   "Bearer token"  default(Bearer <token>)
// @Param        id                 path      int       true   "Product Id"
// @Param        name               formData  string    false  "Product name"
// @Param        description        formData  string    false  "Product description"
// @Param        price              formData  number    false  "Product price"
// @Param        discountPercent    formData  number    false  "Discount percentage"
// @Param        stock              formData  int       false  "Product stock"
// @Param        isFlashSale        formData  bool      false  "Is flash sale"
// @Param        isActive           formData  bool      false  "Is active"
// @Param        isFavourite        formData  bool      false  "Is favourite"
// @Param        image1             formData  file      false  "Product image 1 (JPEG/PNG, max 1MB)"
// @Param        image2             formData  file      false  "Product image 2 (JPEG/PNG, max 1MB)"
// @Param        image3             formData  file      false  "Product image 3 (JPEG/PNG, max 1MB)"
// @Param        image4             formData  file      false  "Product image 4 (JPEG/PNG, max 1MB)"
// @Param        sizeProducts       formData  string    false  "Size Id (comma-separated, e.g., 1,2,3)"
// @Param        productCategories  formData  string    false  "Category Id (comma-separated, e.g., 1,2,3)"
// @Param        productVariants    formData  string    false  "Variant Id (comma-separated, e.g., 1,2,3)"
// @Success      200  {object}  lib.ResponseSuccess  "Product updated successfully"
// @Failure      400  {object}  lib.ResponseError   "Invalid Id format or invalid request body"
// @Failure      404  {object}  lib.ResponseError   "Product not found"
// @Failure      500  {object}  lib.ResponseError   "Internal server error"
// @Router       /admin/products/{id} [patch]
func UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	var bodyUpdate models.ProductRequest
	err = ctx.ShouldBindWith(&bodyUpdate, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// get user id from token
	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// Start transaction
	tx, err := config.DB.Begin(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to start database transaction",
			Error:   err.Error(),
		})
		return
	}
	defer tx.Rollback(context.Background())

	// Update product
	err = models.UpdateDataProduct(tx, id, &bodyUpdate, userIdFromToken.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating product",
			Error:   err.Error(),
		})
		return
	}

	// Process images if any
	fileImages := map[string]*multipart.FileHeader{
		"image1": bodyUpdate.Image1,
		"image2": bodyUpdate.Image2,
		"image3": bodyUpdate.Image3,
		"image4": bodyUpdate.Image4,
	}

	hasFile := false
	for _, file := range fileImages {
		if file != nil {
			hasFile = true
			break
		}
	}

	if hasFile {
		// Delete old images
		err = models.DeleteProductImages(tx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old images",
				Error:   err.Error(),
			})
			return
		}

		var savedImagePaths []string
		for _, file := range fileImages {
			if file == nil {
				continue
			}

			if file.Size > 1<<20 {
				ctx.JSON(http.StatusBadRequest, lib.ResponseError{
					Success: false,
					Message: "Each file size must be less than 1MB",
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

			fileName := fmt.Sprintf("product_%d_%d%s", id, time.Now().UnixNano(), ext)
			savedFilePath := "uploads/products/" + fileName

			os.MkdirAll("uploads/products", 0755)

			if err := ctx.SaveUploadedFile(file, savedFilePath); err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to save uploaded file",
					Error:   err.Error(),
				})
				return
			}

			savedImagePaths = append(savedImagePaths, savedFilePath)
		}

		// Insert new images
		if len(savedImagePaths) > 0 {
			err = models.InsertProductImages(tx, id, savedImagePaths, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product images",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Update sizes
	if strings.TrimSpace(bodyUpdate.SizeProducts) != "" {
		err = models.DeleteProductSizes(tx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old size products",
				Error:   err.Error(),
			})
			return
		}

		sizeProducts := strings.Split(bodyUpdate.SizeProducts, ",")
		var sizeIds []int
		for _, sizeIdStr := range sizeProducts {
			sizeIdStr = strings.TrimSpace(sizeIdStr)
			if sizeIdStr == "" {
				continue
			}
			sizeId, err := strconv.Atoi(sizeIdStr)
			if err != nil {
				continue
			}
			sizeIds = append(sizeIds, sizeId)
		}

		if len(sizeIds) > 0 {
			err = models.InsertProductSizes(tx, id, sizeIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting size product",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Update categories
	if strings.TrimSpace(bodyUpdate.ProductCategories) != "" {
		err = models.DeleteProductCategories(tx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old product categories",
				Error:   err.Error(),
			})
			return
		}

		productCategories := strings.Split(bodyUpdate.ProductCategories, ",")
		var categoryIds []int
		for _, categoryIdStr := range productCategories {
			categoryIdStr = strings.TrimSpace(categoryIdStr)
			if categoryIdStr == "" {
				continue
			}
			categoryId, err := strconv.Atoi(categoryIdStr)
			if err != nil {
				continue
			}
			categoryIds = append(categoryIds, categoryId)
		}

		if len(categoryIds) > 0 {
			err = models.InsertProductCategories(tx, id, categoryIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product category",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Update variants
	if strings.TrimSpace(bodyUpdate.ProductVariants) != "" {
		err = models.DeleteProductVariants(tx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old product variants",
				Error:   err.Error(),
			})
			return
		}

		productVariants := strings.Split(bodyUpdate.ProductVariants, ",")
		var variantIds []int
		for _, variantIdStr := range productVariants {
			variantIdStr = strings.TrimSpace(variantIdStr)
			if variantIdStr == "" {
				continue
			}
			variantId, err := strconv.Atoi(variantIdStr)
			if err != nil {
				continue
			}
			variantIds = append(variantIds, variantId)
		}

		if len(variantIds) > 0 {
			err = models.InsertProductVariants(tx, id, variantIds, userIdFromToken.(int))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product variants",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to commit transaction",
			Error:   err.Error(),
		})
		return
	}

	if err := models.InvalidateProductCache(context.Background()); err != nil {
		fmt.Printf("Warning: Failed to invalidate cache: %v\n", err)
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Product updated successfully",
	})
}

// DeleteProduct    godoc
// @Summary      Delete product
// @Description  Delete product by Id
// @Tags         admin/products
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Product Id"
// @Success      200  {object}  lib.ResponseSuccess  "Product deleted successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Product not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while deleting product data."
// @Router       /admin/products/{id} [delete]
func DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	// delete data product by id
	isSuccess, message, err := models.DeleteDataProduct(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	if !isSuccess {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: message,
		})
		return
	}

	err = models.InvalidateProductCache(context.Background())
	if err != nil {
		fmt.Printf("Warning: Failed to invalidate cache: %v\n", err)
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
	})
}

// ListFavouriteProduct godoc
// @Summary             Get all favourite products
// @Description         Retrieving all favourite products
// @Tags                products
// @Produce             json
// @Param               limit          query     int     false  "Limit of list favourite products"  default(4)  minimum(1)  maximum(20)
// @Success             200            {object}  object{success=bool,message=string,data=[]models.PublicProductResponse,limit=int}  "Successfully retrieved product list"
// @Failure             400            {object}  lib.ResponseError  "Invalid limit"
// @Failure             500            {object}  lib.ResponseError  "Internal server error while fetching or processing product data"
// @Router              /favourite-products [get]
func ListFavouriteProducts(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "4"))

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Limit must be greater than 0",
		})
		return
	}

	if limit > 20 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Limit cannot exceed 20",
		})
		return
	}

	var err error
	var products []models.PublicProductResponse
	rdb := lib.Redis()

	// redis for list favourite products
	cacheListFavouriteProducts, _ := rdb.Get(context.Background(), ctx.Request.RequestURI).Result()
	if cacheListFavouriteProducts == "" {
		products, err = models.GetListFavouriteProducts(limit)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch favourite products from database",
				Error:   err.Error(),
			})
			return
		}

		productsStr, err := json.Marshal(products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization list all products",
				Error:   err.Error(),
			})
			return
		}

		err = rdb.Set(context.Background(), ctx.Request.RequestURI, productsStr, 15*time.Minute).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set list all products to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cacheListFavouriteProducts), &products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal list all products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get all product",
		"data":    products,
		"limit":   limit,
	})
}

// ListProductsPublic  godoc
// @Summary      	   Get list products for public
// @Description  	   Retrieving list products with filter
// @Tags         	   products
// @Produce      	   json
// @Param        	   q   		    query     string   false  "Search name product"
// @Param        	   cat   		query     []string false  "Category of product"
// @Param        	   sort[name]   query     string   false  "Sort by name" Enums(asc, desc)
// @Param        	   sort[price]  query     string   false  "Sort by price" Enums(asc, desc)
// @Param        	   maxPrice   	query     number   false  "Maximum price product"
// @Param        	   minPrice   	query     number   false  "Minimum price product"
// @Param        	   page   		query     int      false  "Page number"  default(1)  minimum(1)
// @Param        	   limit        query     int      false  "Number of items per page"  default(10)  minimum(1)  maximum(50)
// @Success      	   200          {object}  object{success=bool,message=string,data=[]models.PublicProductResponse,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved product list"
// @Failure      	   400          {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure      	   500          {object}  lib.ResponseError  "Internal server error while fetching or processing product data."
// @Router       	   /products [get]
func ListProductsPublic(ctx *gin.Context) {
	search := ctx.Query("q")
	cat := ctx.QueryArray("cat")

	sortField := ""
	sortName := ctx.Query("sort[name]")
	sortPrice := ctx.Query("sort[price]")

	if sortName != "" {
		switch sortName {
		case "asc":
			sortField = "name_asc"
		case "desc":
			sortField = "name_desc"
		}
	} else if sortPrice != "" {
		switch sortPrice {
		case "asc":
			sortField = "price_asc"
		case "desc":
			sortField = "price_desc"
		}
	}

	maxPrice, _ := strconv.ParseFloat(ctx.Query("maxPrice"), 64)
	minPrice, _ := strconv.ParseFloat(ctx.Query("minPrice"), 64)
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'page' must be greater than 0",
		})
		return
	}

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' must be greater than 0",
		})
		return
	}

	if limit > 100 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid pagination parameter: 'limit' cannot exceed 100",
		})
		return
	}

	var totalData int
	var err error
	rdb := lib.Redis()

	// redis for total data products
	cacheTotalDataProducts, _ := rdb.Get(context.Background(), "totalDataProducts").Result()
	if cacheTotalDataProducts == "" {
		totalData, err = models.TotalDataProducts(search)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}
		err = rdb.Set(context.Background(), "totalDataProducts", totalData, 15*time.Minute).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set total products to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cacheTotalDataProducts), &totalData)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal total products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	// redis for list products
	var products []models.PublicProductResponse
	cacheListAllProducts, _ := rdb.Get(context.Background(), ctx.Request.RequestURI).Result()
	if cacheListAllProducts == "" {
		products, err = models.GetListProductsPublic(search, cat, sortField, maxPrice, minPrice, limit, page)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}

		productsStr, err := json.Marshal(products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization list all products",
				Error:   err.Error(),
			})
			return
		}

		err = rdb.Set(context.Background(), ctx.Request.RequestURI, productsStr, 15*time.Minute).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set list all products to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cacheListAllProducts), &products)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal list all products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/products", scheme, host)

	var next any
	var prev any

	if totalData == 0 {
		page = 0
		next = nil
		prev = nil
	} else if page == 1 && totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	} else if page == totalPage && totalPage > 1 {
		next = nil
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else if totalPage > 1 {
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	} else {
		next = nil
		prev = nil
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get all product",
		"data":    products,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
			"next":        next,
			"prev":        prev,
		},
	})
}

// DetailProduct   godoc
// @Summary        Get product by Id
// @Description    Retrieving product data based on Id for public
// @Tags           products
// @Accept 		   x-www-form-urlencoded
// @Produce        json
// @Param          id   			path    int     true  "product Id"
// @Success        200  {object}  lib.ResponseSuccess{data=models.PublicProductDetailResponse}  "Successfully retrieved product"
// @Failure        400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure        404  {object}  lib.ResponseError  "Product not found"
// @Failure        500  {object}  lib.ResponseError  "Internal server error while fetching products from database"
// @Router         /products/{id} [get]
func DetailProductPublic(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	rdb := lib.Redis()

	// redis for detail product
	var product models.PublicProductDetailResponse
	cache, _ := rdb.Get(context.Background(), ctx.Request.RequestURI).Result()
	if cache == "" {
		product, message, err := models.GetDetailProductPublic(id)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if message == "Product not found" {
				statusCode = http.StatusNotFound
			}
			ctx.JSON(statusCode, lib.ResponseError{
				Success: false,
				Message: message,
				Error:   err.Error(),
			})
			return
		}

		productStr, err := json.Marshal(product)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization detail product",
				Error:   err.Error(),
			})
			return
		}

		err = rdb.Set(context.Background(), ctx.Request.RequestURI, productStr, 60*time.Second).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to set data product to cache",
				Error:   err.Error(),
			})
			return
		}
	} else {
		err = json.Unmarshal([]byte(cache), &product)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to unmarshal data products from cache",
				Error:   err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get product",
		Data:    product,
	})
}
