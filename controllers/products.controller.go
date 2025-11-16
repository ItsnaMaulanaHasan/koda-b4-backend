package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/redis/go-redis/v9"
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

	cacheTotalDataProducts, err := rdb.Get(context.Background(), totalCacheKey).Result()
	if err == redis.Nil || cacheTotalDataProducts == "" {
		// cache miss - ambil dari DB
		totalData, err = models.TotalDataProducts(search)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}

		// simpan ke cache
		cacheErr := rdb.Set(context.Background(), totalCacheKey, totalData, 15*time.Minute).Err()
		if cacheErr != nil {
			log.Printf("Failed to set total cache for key %s: %v", totalCacheKey, cacheErr)
		}
	} else if err != nil {
		// redis error - fallback ke DB
		log.Printf("Redis error for key %s: %v", totalCacheKey, err)

		totalData, err = models.TotalDataProducts(search)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// cache hit - unmarshal
		err = json.Unmarshal([]byte(cacheTotalDataProducts), &totalData)
		if err != nil {
			// cache rusak - hapus dan ambil dari DB
			log.Printf("Failed to unmarshal total cache for key %s: %v", totalCacheKey, err)
			rdb.Del(context.Background(), totalCacheKey)

			totalData, err = models.TotalDataProducts(search)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to count total products in database",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	listCacheKey := fmt.Sprintf("products:list:page:%d:limit:%d:search:%s", page, limit, search)

	var products []models.AdminProductResponse
	cacheListAllProducts, err := rdb.Get(context.Background(), listCacheKey).Result()
	if err == redis.Nil || cacheListAllProducts == "" {
		// cache miss - ambil dari DB
		products, err = models.GetListProductsAdmin(search, page, limit)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}

		// simpan ke cache
		productsStr, marshalErr := json.Marshal(products)
		if marshalErr != nil {
			log.Printf("Failed to marshal products: %v", marshalErr)
		} else {
			cacheErr := rdb.Set(context.Background(), listCacheKey, productsStr, 15*time.Minute).Err()
			if cacheErr != nil {
				log.Printf("Failed to set list cache for key %s: %v", listCacheKey, cacheErr)
			}
		}
	} else if err != nil {
		// redis error - fallback ke DB
		log.Printf("Redis error for key %s: %v", listCacheKey, err)

		products, err = models.GetListProductsAdmin(search, page, limit)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// cache hit - unmarshal
		err = json.Unmarshal([]byte(cacheListAllProducts), &products)
		if err != nil {
			// cache rusak - hapus dan ambil dari DB
			log.Printf("Failed to unmarshal list cache for key %s: %v", listCacheKey, err)
			rdb.Del(context.Background(), listCacheKey)

			products, err = models.GetListProductsAdmin(search, page, limit)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to fetch products from database",
					Error:   err.Error(),
				})
				return
			}
		}
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
// @Param          		id   			 path    int     true  "product Id"
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
	cacheKey := fmt.Sprintf("products:detail:id:%d", id)

	// redis for detail product
	var product models.AdminProductResponse
	var message string

	cache, err := rdb.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil || cache == "" {
		product, message, err = models.GetDetailProductAdmin(id)
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

		productStr, marshalErr := json.Marshal(product)
		if marshalErr != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to serialization detail product",
				Error:   marshalErr.Error(),
			})
			return
		}

		cacheErr := rdb.Set(context.Background(), cacheKey, productStr, 15*time.Minute).Err()
		if cacheErr != nil {
			log.Printf("Failed to set cache for key %s: %v", cacheKey, cacheErr)
		}
	} else if err != nil {
		log.Printf("Redis error for key %s: %v", cacheKey, err)

		product, message, err = models.GetDetailProductAdmin(id)
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
	} else {
		err = json.Unmarshal([]byte(cache), &product)
		if err != nil {
			log.Printf("Failed to unmarshal cache for key %s: %v", cacheKey, err)
			rdb.Del(context.Background(), cacheKey)

			product, message, err = models.GetDetailProductAdmin(id)
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
// @Param        fileImages         formData  file      false  "Product images (4 files required, JPEG/PNG, max 1MB each)"
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

	// get uploaded files
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to parse multipart form",
			Error:   err.Error(),
		})
		return
	}

	files := form.File["images"]

	// validate amount of uploaded images
	if len(files) > 4 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "There must be no more than 4 product images uploaded",
		})
		return
	}

	// validate each file
	allowedTypes := map[string]bool{
		"image/jpg":  true,
		"image/jpeg": true,
		"image/png":  true,
	}
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	for i, file := range files {
		// check file size
		if file.Size > 1<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: fmt.Sprintf("Image %d size must be less than 1MB (got %.2f MB)", i+1, float64(file.Size)/(1<<20)),
			})
			return
		}

		// check content type
		contentType := file.Header.Get("Content-Type")
		if !allowedTypes[contentType] {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: fmt.Sprintf("Image %d has invalid type. Only JPEG and PNG are allowed", i+1),
			})
			return
		}

		// check file extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedExt[ext] {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: fmt.Sprintf("Image %d has invalid extension. Only JPG and PNG are allowed", i+1),
			})
			return
		}
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

	// start transaction
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

	// insert product
	err = models.InsertDataProduct(tx, &bodyCreate, userIdFromToken.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new product",
			Error:   err.Error(),
		})
		return
	}

	// process and save images
	var savedImagePaths []string
	uploadDir := "uploads/products"
	os.MkdirAll(uploadDir, 0755)

	for i, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		fileName := fmt.Sprintf("product_%d_img%d_%d%s", bodyCreate.Id, i+1, time.Now().UnixNano(), ext)
		savedFilePath := filepath.Join(uploadDir, fileName)

		err := ctx.SaveUploadedFile(file, savedFilePath)
		if err != nil {
			// clean up already saved files
			for _, path := range savedImagePaths {
				os.Remove(path)
			}
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: fmt.Sprintf("Failed to save image %d", i+1),
				Error:   err.Error(),
			})
			return
		}

		savedImagePaths = append(savedImagePaths, savedFilePath)
	}

	// insert images
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

	// insert sizes
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

	// insert categories
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

	// insert variants
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

	// commit transaction
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
// @Param        fileImages         formData  file      false  "Product images (up to 4 files, JPEG/PNG, max 1MB each)"
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

	// get uploaded files (if any)
	form, err := ctx.MultipartForm()
	var files []*multipart.FileHeader
	if err == nil && form != nil {
		files = form.File["images"]
	}

	// validate files if uploaded
	if len(files) > 0 {
		// validate maximum 4 images
		if len(files) > 4 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: fmt.Sprintf("Maximum 4 images allowed, but got %d", len(files)),
			})
			return
		}

		// validate each file
		allowedTypes := map[string]bool{
			"image/jpg":  true,
			"image/jpeg": true,
			"image/png":  true,
		}
		allowedExt := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		for i, file := range files {
			// check file size
			if file.Size > 1<<20 {
				ctx.JSON(http.StatusBadRequest, lib.ResponseError{
					Success: false,
					Message: fmt.Sprintf("Image %d size must be less than 1MB (got %.2f MB)", i+1, float64(file.Size)/(1<<20)),
				})
				return
			}

			// check content type
			contentType := file.Header.Get("Content-Type")
			if !allowedTypes[contentType] {
				ctx.JSON(http.StatusBadRequest, lib.ResponseError{
					Success: false,
					Message: fmt.Sprintf("Image %d has invalid type. Only JPEG and PNG are allowed", i+1),
				})
				return
			}

			// check file extension
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !allowedExt[ext] {
				ctx.JSON(http.StatusBadRequest, lib.ResponseError{
					Success: false,
					Message: fmt.Sprintf("Image %d has invalid extension. Only JPG and PNG are allowed", i+1),
				})
				return
			}
		}
	}

	// start transaction
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

	// update product
	commandTag, err := models.UpdateDataProduct(tx, id, &bodyUpdate, userIdFromToken.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating product",
			Error:   err.Error(),
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Product not found",
		})
		return
	}

	if len(files) > 0 {
		// delete old images from database
		err = models.DeleteProductImages(tx, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old images",
				Error:   err.Error(),
			})
			return
		}

		// process and save new images
		var savedImagePaths []string
		uploadDir := "uploads/products"
		os.MkdirAll(uploadDir, 0755)

		for i, file := range files {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			fileName := fmt.Sprintf("product_%d_img%d_%d%s", id, i+1, time.Now().UnixNano(), ext)
			savedFilePath := filepath.Join(uploadDir, fileName)

			err := ctx.SaveUploadedFile(file, savedFilePath)
			if err != nil {
				// clean up already saved files
				for _, path := range savedImagePaths {
					os.Remove(path)
				}
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: fmt.Sprintf("Failed to save image %d", i+1),
					Error:   err.Error(),
				})
				return
			}

			savedImagePaths = append(savedImagePaths, savedFilePath)
		}

		// insert new images
		if len(savedImagePaths) > 0 {
			err = models.InsertProductImages(tx, id, savedImagePaths, userIdFromToken.(int))
			if err != nil {
				// clean up saved files on error
				for _, path := range savedImagePaths {
					os.Remove(path)
				}
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product images",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	// update sizes
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

	// update categories
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

	// update variants
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

	// commit transaction
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

	// Ubah cache key untuk include semua filter
	totalCacheKey := fmt.Sprintf("productsPublic:total:q:%s:cat:%v:maxPrice:%f:minPrice:%f",
		search, cat, maxPrice, minPrice)

	cacheTotalDataProducts, err := rdb.Get(context.Background(), totalCacheKey).Result()
	if err == redis.Nil || cacheTotalDataProducts == "" {
		// cache miss - ambil dari DB dengan filter lengkap
		totalData, err = models.TotalDataProductsPublic(search, cat, maxPrice, minPrice)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}

		// simpan ke cache
		cacheErr := rdb.Set(context.Background(), totalCacheKey, totalData, 15*time.Minute).Err()
		if cacheErr != nil {
			log.Printf("Failed to set total cache for key %s: %v", totalCacheKey, cacheErr)
		}
	} else if err != nil {
		// redis error - fallback ke DB
		log.Printf("Redis error for key %s: %v", totalCacheKey, err)

		totalData, err = models.TotalDataProductsPublic(search, cat, maxPrice, minPrice)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to count total products in database",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// cache hit - unmarshal
		err = json.Unmarshal([]byte(cacheTotalDataProducts), &totalData)
		if err != nil {
			// cache rusak - hapus dan ambil dari DB
			log.Printf("Failed to unmarshal total cache for key %s: %v", totalCacheKey, err)
			rdb.Del(context.Background(), totalCacheKey)

			totalData, err = models.TotalDataProductsPublic(search, cat, maxPrice, minPrice)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to count total products in database",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	listCacheKey := fmt.Sprintf("productsPublic:list:page:%d:limit:%d:q:%s:cat:%v:sort:%s:maxPrice:%f:minPrice:%f",
		page, limit, search, cat, sortField, maxPrice, minPrice)

	var products []models.PublicProductResponse
	cacheListAllProducts, err := rdb.Get(context.Background(), listCacheKey).Result()
	if err == redis.Nil || cacheListAllProducts == "" {
		// cache miss - ambil dari DB
		products, err = models.GetListProductsPublic(search, cat, sortField, maxPrice, minPrice, limit, page)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}

		// simpan ke cache
		productsStr, marshalErr := json.Marshal(products)
		if marshalErr != nil {
			log.Printf("Failed to marshal products: %v", marshalErr)
		} else {
			cacheErr := rdb.Set(context.Background(), listCacheKey, productsStr, 15*time.Minute).Err()
			if cacheErr != nil {
				log.Printf("Failed to set list cache for key %s: %v", listCacheKey, cacheErr)
			}
		}
	} else if err != nil {
		// redis error - fallback ke DB
		log.Printf("Redis error for key %s: %v", listCacheKey, err)

		products, err = models.GetListProductsPublic(search, cat, sortField, maxPrice, minPrice, limit, page)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to fetch products from database",
				Error:   err.Error(),
			})
			return
		}
	} else {
		// cache hit - unmarshal
		err = json.Unmarshal([]byte(cacheListAllProducts), &products)
		if err != nil {
			// cache rusak - hapus dan ambil dari DB
			log.Printf("Failed to unmarshal list cache for key %s: %v", listCacheKey, err)
			rdb.Del(context.Background(), listCacheKey)

			products, err = models.GetListProductsPublic(search, cat, sortField, maxPrice, minPrice, limit, page)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to fetch products from database",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/products", scheme, host)

	var next any
	var prev any

	if page == 1 && totalPage > 1 {
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
