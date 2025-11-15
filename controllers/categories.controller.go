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
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
)

// GetAllCategory    godoc
// @Summary      	 Get all categories
// @Description  	 Retrieving all category data with pagination support
// @Tags         	 admin/categories
// @Produce      	 json
// @Security     	 BearerAuth
// @Param        	 Authorization  header    string  true   "Bearer token"              default(Bearer <token>)
// @Param        	 page           query     int     false  "Page number"               default(1)   minimum(1)
// @Param        	 limit          query     int     false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        	 search         query     string  false  "Search value"
// @Success      	 200  {object}  object{success=bool,message=string,data=[]models.Category,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved category list"
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

	// get total data categories
	totalData, err := models.GetTotalDataCategories(search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total categories in database",
			Error:   err.Error(),
		})
		return
	}

	// get list all categories
	categories, message, err := models.GetListAllCategories(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// get total page
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
	baseURL := fmt.Sprintf("%s://%s/admin/categories", scheme, host)

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
		"message": message,
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
// @Tags         admin/categories
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Category Id"
// @Success      200  {object}  lib.ResponseSuccess{data=models.Category}  "Successfully retrieved category"
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

	category, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Category])
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

// CreateCategory    godoc
// @Summary      Create new category
// @Description  Create a new category with a unique name
// @Tags         admin/categories
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        name           formData  string  true  "Category name"
// @Success      201  {object}  lib.ResponseSuccess{data=models.Category}  "Category created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      409  {object}  lib.ResponseError  "Category name already exists"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while creating category"
// @Router       /admin/categories [post]
func CreateCategory(ctx *gin.Context) {
	var bodyCreateCategory models.Category
	err := ctx.ShouldBindWith(&bodyCreateCategory, binding.Form)
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
		"SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)", bodyCreateCategory.Name,
	).Scan(&exists)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking category name uniqueness",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Category name already exists",
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
		`INSERT INTO categories (name, created_by, updated_by)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		bodyCreateCategory.Name,
		userIdFromToken,
		userIdFromToken,
	).Scan(&bodyCreateCategory.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new category",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: "Category created successfully",
		Data: models.Category{
			Id:   bodyCreateCategory.Id,
			Name: bodyCreateCategory.Name,
		},
	})
}

// UpdateCategory    godoc
// @Summary      Update category
// @Description  Updating category data based on Id
// @Tags         admin/categories
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path      int     true  "Category Id"
// @Param        name           formData  string  true  "Category name"
// @Success      200  {object}  lib.ResponseSuccess  "Category updated successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format or invalid request body"
// @Failure      404  {object}  lib.ResponseError  "Category not found"
// @Failure      409  {object}  lib.ResponseError  "Category name already exists"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while updating category data"
// @Router       /admin/categories/{id} [patch]
func UpdateCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	bodyUpdate := ctx.PostForm("name")
	if bodyUpdate == "" {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Status is required",
		})
		return
	}

	var exists bool
	err = config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND id != $2)", bodyUpdate, id,
	).Scan(&exists)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking category name uniqueness",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Category name already exists",
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

	commandTag, err := config.DB.Exec(
		context.Background(),
		`UPDATE categories 
		 SET name = COALESCE(NULLIF($1, ''), name),
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		bodyUpdate,
		userIdFromToken,
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating category",
			Error:   err.Error(),
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Category updated successfully",
	})
}

// DeleteCategory    godoc
// @Summary      Delete category
// @Description  Delete category by Id
// @Tags         admin/categories
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "Category Id"
// @Success      200  {object}  lib.ResponseSuccess  "Category deleted successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "Category not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while deleting category data"
// @Router       /admin/categories/{id} [delete]
func DeleteCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while deleting category data",
			Error:   err.Error(),
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Category not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Category deleted successfully",
	})
}
