package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
)

// GetAllUser godoc
// @Summary      Get all users
// @Description  Retrieving all user data with pagination support
// @Tags         users
// @Produce      json
// @Param        page   query     int  false  "Page number"  default(1)  minimum(1)
// @Param        limit  query     int  false  "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Success      200    {object}  object{success=bool,message=string,data=[]models.User,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int}}  "Success get all users"
// @Failure      400    {object}  lib.Response  "Invalid pagination parameters"
// @Failure      500    {object}  lib.Response  "Invalid query"
// @Router       /users [get]
func GetAllUser(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Page must be greater than 0",
		})
		return
	}

	if limit < 1 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Limit must be greater than 0",
		})
		return
	}

	if limit > 100 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Limit cannot exceed 100",
		})
		return
	}

	var totalData int
	err := config.DB.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users").Scan(&totalData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to count users",
		})
		return
	}

	offset := (page - 1) * limit
	rows, err := config.DB.Query(context.Background(),
		"SELECT id, first_name, last_name, email, role FROM users LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to query users",
		})
		return
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to collect users",
		})
		return
	}

	totalPage := (totalData + limit - 1) / limit
	if page > totalPage && totalData > 0 {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Page is out of range",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Success get all user",
		"data":    users,
		"meta": gin.H{
			"currentPage": page,
			"perPage":     limit,
			"totalData":   totalData,
			"totalPages":  totalPage,
		},
	})
}

// GetUserById godoc
// @Summary      Get user by Id
// @Description  Retrieving user data based on Id
// @Tags         users
// @Accept 		 x-www-form-urlencoded
// @Produce      json
// @Param        id   path      int  true  "User Id"
// @Success      200  {object}  lib.Response{data=models.User}  "Success get user"
// @Failure      400  {object}  lib.Response  "Invalid Id format"
// @Failure      404  {object}  lib.Response  "User not found"
// @Failure      500  {object}  lib.Response  "Failed to query users"
// @Router       /users/{id} [get]
func GetUserById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Invalid Id format",
		})
		return
	}

	var foundUser *models.User
	err = config.DB.QueryRow(context.Background(),
		"SELECT id, first_name, last_name, email, role FROM users WHERE id = $1", id).Scan(&foundUser)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to query users",
		})
		return
	}

	if foundUser != nil {
		ctx.JSON(http.StatusOK, lib.Response{
			Success: true,
			Message: "Success get user",
			Data:    foundUser,
		})
	} else {
		ctx.JSON(http.StatusNotFound, lib.Response{
			Success: false,
			Message: "User not found",
		})
	}
}

// CreateUser godoc
// @Summary      Create new user
// @Description  Create a new user with a unique username and email
// @Tags         users
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        user      formData  models.User true "User registration data"
// @Param        role      formData  string false "User role" default("customer")
// @Success      200       {object}  lib.Response{data=models.User}  "User created successfully"
// @Failure      400       {object}  lib.Response  "Invalid request body or hash password failed"
// @Failure      409       {object}  lib.Response  "Email or username already exists"
// @Failure      500       {object}  lib.Response  "Failed to insert user"
// @Router       /users [post]
func CreateUser(ctx *gin.Context) {
	var body models.User
	err := ctx.ShouldBindWith(&body, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: true,
			Message: err.Error(),
		})
		return
	}

	rows, err := config.DB.Query(context.Background(),
		"SELECT id, first_name, last_name, email, role FROM users")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to query users",
		})
		return
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to get users",
		})
		return
	}

	for _, user := range users {
		if user.Email == body.Email {
			ctx.JSON(http.StatusConflict, lib.Response{
				Success: false,
				Message: "Email already registered",
			})
			return
		}
	}

	hashPassword, err := lib.HashPassword(body.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.Response{
			Success: false,
			Message: "Hash password failed",
		})
		return
	}

	body.Password = string(hashPassword)

	_, err = config.DB.Exec(ctx,
		`INSERT INTO users (first_name, last_name, email, role, password) VALUES ($1, $2, $3, $4, $5)`, body.FirstName, body.LastName, body.Email, body.Role, body.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.Response{
			Success: false,
			Message: "Failed to insert user",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.Response{
		Success: true,
		Message: "User created successfully",
		Data:    body,
	})
}
