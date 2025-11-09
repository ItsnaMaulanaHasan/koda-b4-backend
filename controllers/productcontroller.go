package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetAllProduct godoc
// @Summary      Get all product
// @Description  Retrieving all product data with pagination support
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization    header string true "Bearer token" default(Bearer <token>)
// @Param        page   query     int    false  "Page number"  default(1)  minimum(1)
// @Param        limit  query     int    false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Success      200    {object}  object{success=bool,message=string,data=[]models.Product,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int}}  "Successfully retrieved user list."
// @Failure      400    {object}  lib.ResponseError  "Invalid pagination parameters or page out of range."
// @Failure      500    {object}  lib.ResponseError  "Internal server error while fetching or processing user data."
// @Router       /admin/products [get]
func GetAllProduct(ctx *gin.Context) {
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
	err := config.DB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM products").Scan(&totalData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total products in database",
			Error:   err.Error(),
		})
		return
	}

	offset := (page - 1) * limit

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
					COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS image,
					COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
					COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
				FROM products p
				LEFT JOIN product_images pi ON pi.product_id = p.id
				LEFT JOIN size_products sp ON sp.product_id = p.id
				LEFT JOIN sizes s ON s.id = sp.size_id
				LEFT JOIN product_categories pc ON pc.product_id = p.id
				LEFT JOIN categories c ON c.id = pc.category_id
				GROUP BY p.id
				ORDER BY p.id ASC
				LIMIT $1 OFFSET $2;`

	rows, err := config.DB.Query(context.Background(), query, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch products from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Product])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process product data from database",
			Error:   err.Error(),
		})
		return
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Page is out of range",
		})
		return
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
		},
	})
}
