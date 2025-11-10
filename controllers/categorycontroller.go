package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// GetAllCategory    godoc
// @Summary      	 Get all categories
// @Description  	 Retrieving all category data with pagination support
// @Tags         	 categories
// @Produce      	 json
// @Security     	 BearerAuth
// @Param        	 Authorization  header    string  true   "Bearer token"              default(Bearer <token>)
// @Param        	 page           query     int     false  "Page number"               default(1)   minimum(1)
// @Param        	 limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        	 search         query     string  false  "Search value"
// @Success      	 200  {object}  object{success=bool,message=string,data=[]models.CategoryResponse,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved category list"
// @Failure      	 400  {object}  lib.ResponseError  "Invalid pagination parameters or page out of range"
// @Failure      	 500  {object}  lib.ResponseError  "Internal server error while fetching or processing category data"
// @Router       	 /admin/categories [get]
func GetAllCategory(ctx *gin.Context) {
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
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM categories WHERE name ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM categories`).Scan(&totalData)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total categories in database",
			Error:   err.Error(),
		})
		return
	}

	offset := (page - 1) * limit
	var rows pgx.Rows
	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM categories
			WHERE name ILIKE $3
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT id, name, created_at, updated_at
			FROM categories
			ORDER BY id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch categories from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.CategoryResponse])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process category data from database",
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

	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/admin/categories", scheme, host)

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
		"message": "Success get all categories",
		"data":    categories,
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

// GetCategoryById   godoc
// @Summary      Get category by Id
// @Description  Retrieving category data based on Id
// @Tags         categories
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Category Id"
// @Success      200  {object}  lib.ResponseSuccess{data=models.CategoryResponse}  "Successfully retrieved category"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Category not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while fetching category from database"
// @Router       /admin/categories/{id} [get]
func GetCategoryById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	rows, err := config.DB.Query(context.Background(),
		`SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id = $1`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch category from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	category, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.CategoryResponse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Category not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process category data",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get category",
		Data:    category,
	})
}
