package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
)

// GetAllProduct godoc
// @Summary      Get all product
// @Description  Retrieving all product data with pagination support
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true   "Bearer token" default(Bearer <token>)
// @Param        page   		query     int     false  "Page number"  default(1)  minimum(1)
// @Param        limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        search         query     string  false  "Search value"
// @Success      200            {object}  object{success=bool,message=string,data=[]models.Product,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved product list"
// @Failure      400            {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure      500            {object}  lib.ResponseError  "Internal server error while fetching or processing product data."
// @Router       /admin/products [get]
func ListAllProducts(ctx *gin.Context) {
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

	cacheTotalDataProducts, _ := lib.Redis().Get(context.Background(), "totalDataProduct").Result()
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
		err = lib.Redis().Set(context.Background(), "totalDataProducts", totalData, 0).Err()
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

	var products []models.Product
	cacheListAllProducts, _ := lib.Redis().Get(context.Background(), ctx.Request.RequestURI).Result()
	if cacheListAllProducts == "" {
		products, err = models.GetListAllProducts(search, page, limit)
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

		err = lib.Redis().Set(context.Background(), ctx.Request.RequestURI, productsStr, 60*time.Second).Err()
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
	baseURL := fmt.Sprintf("%s://%s/admin/users", scheme, host)

	var next any
	var prev any
	switch page {
	case 1:
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	case totalPage:
		next = nil
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	default:
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page-1, limit)
	}

	if totalData == 0 {
		page = 0
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

// GetProductById   godoc
// @Summary         Get product by Id
// @Description     Retrieving product data based on Id
// @Tags         	products
// @Accept 		 	x-www-form-urlencoded
// @Produce      	json
// @Security     	BearerAuth
// @Param        	Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        	id   			path    int     true  "product Id"
// @Success      	200  {object}  lib.ResponseSuccess{data=models.Product}  "Successfully retrieved product"
// @Failure      	400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      	404  {object}  lib.ResponseError  "Product not found"
// @Failure      	500  {object}  lib.ResponseError  "Internal server error while fetching products from database"
// @Router       	/admin/products/{id} [get]
func DetailProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	query := `SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				p.is_active,
				COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id 
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch product from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	product, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Product])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Product not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process product data",
			Error:   err.Error(),
		})
		return
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
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization       header    string    true   "Bearer token"  default(Bearer <token>)
// @Param        name                formData  string    true   "Product name"
// @Param        description         formData  string    true   "Product description"
// @Param        price               formData  number    true   "Product price"
// @Param        discount_percent    formData  number    false  "Discount percentage" default(0.00)
// @Param        rating			     formData  number    false  "Product rating" default(5)
// @Param        stock               formData  int       true   "Product stock"
// @Param        is_flash_sale       formData  bool      false  "Is flash sale"  default(false)
// @Param        is_active           formData  bool      false  "Is active"  default(true)
// @Param        image1              formData  file      false  "Product image 1 (JPEG/PNG, max 1MB)"
// @Param        image2              formData  file      false  "Product image 2 (JPEG/PNG, max 1MB)"
// @Param        image3              formData  file      false  "Product image 3 (JPEG/PNG, max 1MB)"
// @Param        image4              formData  file      false  "Product image 4 (JPEG/PNG, max 1MB)"
// @Param        size_products       formData  string    false  "Size Id (comma-separated, e.g., 1,2,3)"
// @Param        product_categories  formData  string    false  "Category Id (comma-separated, e.g., 1,2,3)"
// @Success      201  {object}  lib.ResponseSuccess{data=models.Product}  "Product created successfully"
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

	var exists bool
	err = config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)", bodyCreate.Name,
	).Scan(&exists)
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

	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	err = config.DB.QueryRow(
		context.Background(),
		`INSERT INTO products (name, description, price, discount_percent, rating, is_flash_sale, stock, is_active, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		bodyCreate.Name,
		bodyCreate.Description,
		bodyCreate.Price,
		bodyCreate.DiscountPercent,
		bodyCreate.Rating,
		bodyCreate.IsFlashSale,
		bodyCreate.Stock,
		bodyCreate.IsActive,
		userIdFromToken,
		userIdFromToken,
	).Scan(&bodyCreate.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new product",
			Error:   err.Error(),
		})
		return
	}

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
		if contentType != "image/jpeg" && contentType != "image/png" {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "Only JPEG and PNG files are allowed",
			})
			return
		}

		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("product_%d_%d%s", bodyCreate.Id, time.Now().UnixNano(), ext)
		savedFilePath := "uploads/products/" + fileName

		if err := ctx.SaveUploadedFile(file, savedFilePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to save uploaded file",
				Error:   err.Error(),
			})
			return
		}

		_, err = config.DB.Exec(
			context.Background(),
			`INSERT INTO product_images (image, product_id, created_by, updated_by)
				 VALUES ($1, $2, $3, $4)`,
			savedFilePath,
			bodyCreate.Id,
			userIdFromToken,
			userIdFromToken,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while inserting product image",
				Error:   err.Error(),
			})
			return
		}
		savedImagePaths = append(savedImagePaths, savedFilePath)
	}
	bodyCreate.Images = savedImagePaths

	if strings.TrimSpace(bodyCreate.SizeProducts) != "" {
		sizeProducts := strings.Split(bodyCreate.SizeProducts, ",")
		for _, sizeIdStr := range sizeProducts {
			sizeIdStr = strings.TrimSpace(sizeIdStr)
			if sizeIdStr == "" {
				continue
			}

			sizeId, err := strconv.Atoi(sizeIdStr)
			if err != nil {
				continue
			}

			_, err = config.DB.Exec(
				context.Background(),
				`INSERT INTO size_products (product_id, size_id, created_by, updated_by)
				 VALUES ($1, $2, $3, $4)`,
				bodyCreate.Id,
				sizeId,
				userIdFromToken,
				userIdFromToken,
			)
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

	if strings.TrimSpace(bodyCreate.ProductCategories) != "" {
		productCategories := strings.SplitSeq(bodyCreate.ProductCategories, ",")

		for categoryIdStr := range productCategories {
			categoryIdStr = strings.TrimSpace(categoryIdStr)
			if categoryIdStr == "" {
				continue
			}

			categoryId, err := strconv.Atoi(categoryIdStr)
			if err != nil {
				continue
			}

			_, err = config.DB.Exec(
				context.Background(),
				`INSERT INTO product_categories (product_id, category_id, created_by, updated_by)
				 VALUES ($1, $2, $3, $4)`,
				bodyCreate.Id,
				categoryId,
				userIdFromToken,
				userIdFromToken,
			)
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

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: "Product created successfully",
		Data:    bodyCreate,
	})
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Updating product data based on Id
// @Tags         products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization       header    string    true   "Bearer token"  default(Bearer <token>)
// @Param        id                  path      int       true   "Product Id"
// @Param        name                formData  string    false  "Product name"
// @Param        description         formData  string    false  "Product description"
// @Param        price               formData  number    false  "Product price"
// @Param        discount_percent    formData  number    false  "Discount percentage"
// @Param        stock               formData  int       false  "Product stock"
// @Param        is_flash_sale       formData  bool      false  "Is flash sale"
// @Param        is_active           formData  bool      false  "Is active"
// @Param        image1              formData  file      false  "Product image 1 (JPEG/PNG, max 1MB)"
// @Param        image2              formData  file      false  "Product image 2 (JPEG/PNG, max 1MB)"
// @Param        image3              formData  file      false  "Product image 3 (JPEG/PNG, max 1MB)"
// @Param        image4              formData  file      false  "Product image 4 (JPEG/PNG, max 1MB)"
// @Param        size_products       formData  string    false  "Size Id (comma-separated, e.g., 1,2,3)"
// @Param        product_categories  formData  string    false  "Category Id (comma-separated, e.g., 1,2,3)"
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

	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE products 
		 SET name             = COALESCE(NULLIF($1, ''), name),
		     description      = COALESCE(NULLIF($2, ''), description),
		     price            = COALESCE($3, price),
		     discount_percent = COALESCE($4, discount_percent),
		     stock            = COALESCE($5, stock),
		     is_flash_sale    = COALESCE($6, is_flash_sale),
		     is_active        = COALESCE($7, is_active),
		     updated_by       = $8,
		     updated_at       = NOW()
		 WHERE id = $9`,
		bodyUpdate.Name,
		bodyUpdate.Description,
		bodyUpdate.Price,
		bodyUpdate.DiscountPercent,
		bodyUpdate.Stock,
		bodyUpdate.IsFlashSale,
		bodyUpdate.IsActive,
		userIdFromToken,
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating product",
			Error:   err.Error(),
		})
		return
	}

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
		_, err = config.DB.Exec(
			context.Background(),
			`DELETE FROM product_images WHERE product_id = $1`,
			id,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old images",
				Error:   err.Error(),
			})
			return
		}

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
			if contentType != "image/jpeg" && contentType != "image/png" {
				ctx.JSON(http.StatusBadRequest, lib.ResponseError{
					Success: false,
					Message: "Only JPEG and PNG files are allowed",
				})
				return
			}

			ext := filepath.Ext(file.Filename)
			fileName := fmt.Sprintf("product_%d_%d%s", id, time.Now().UnixNano(), ext)
			savedFilePath := "uploads/products/" + fileName

			if err := ctx.SaveUploadedFile(file, savedFilePath); err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Failed to save uploaded file",
					Error:   err.Error(),
				})
				return
			}

			_, err = config.DB.Exec(
				context.Background(),
				`INSERT INTO product_images (image, product_id, created_by, updated_by)
						 VALUES ($1, $2, $3, $4)`,
				savedFilePath,
				id,
				userIdFromToken,
				userIdFromToken,
			)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
					Success: false,
					Message: "Internal server error while inserting product image",
					Error:   err.Error(),
				})
				return
			}
		}
	}

	if strings.TrimSpace(bodyUpdate.SizeProducts) != "" {
		sizeProducts := strings.Split(bodyUpdate.SizeProducts, ",")
		_, err = config.DB.Exec(
			context.Background(),
			`DELETE FROM size_products WHERE product_id = $1`,
			id,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old size products",
				Error:   err.Error(),
			})
			return
		}

		for _, sizeIdStr := range sizeProducts {
			sizeIdStr = strings.TrimSpace(sizeIdStr)
			if sizeIdStr == "" {
				continue
			}

			sizeId, err := strconv.Atoi(sizeIdStr)
			if err != nil {
				continue
			}

			_, err = config.DB.Exec(
				context.Background(),
				`INSERT INTO size_products (product_id, size_id, created_by, updated_by)
				 VALUES ($1, $2, $3, $4)`,
				id,
				sizeId,
				userIdFromToken,
				userIdFromToken,
			)
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

	if strings.TrimSpace(bodyUpdate.ProductCategories) != "" {
		productCategories := strings.Split(bodyUpdate.ProductCategories, ",")
		_, err = config.DB.Exec(
			context.Background(),
			`DELETE FROM product_categories WHERE product_id = $1`,
			id,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Internal server error while deleting old product categories",
				Error:   err.Error(),
			})
			return
		}

		for _, categoryIdStr := range productCategories {
			categoryIdStr = strings.TrimSpace(categoryIdStr)
			if categoryIdStr == "" {
				continue
			}

			categoryId, err := strconv.Atoi(categoryIdStr)
			if err != nil {
				continue
			}

			_, err = config.DB.Exec(
				context.Background(),
				`INSERT INTO product_categories (product_id, category_id, created_by, updated_by)
				 VALUES ($1, $2, $3, $4)`,
				id,
				categoryId,
				userIdFromToken,
				userIdFromToken,
			)
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

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Product updated successfully",
	})
}

// DeleteUser    godoc
// @Summary      Delete product
// @Description  Delete product by Id
// @Tags         products
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

	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while deleting product data",
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

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Product deleted successfully",
	})
}
