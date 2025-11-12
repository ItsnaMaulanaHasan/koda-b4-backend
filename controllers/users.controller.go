package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
)

// ListUsers     godoc
// @Summary      Get list users
// @Description  Retrieving list users with pagination support
// @Tags         admin/users
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header 	  string  true       "Bearer token"              default(Bearer <token>)
// @Param        page   		query     int     false      "Page number"  			 default(1)   minimum(1)
// @Param        limit  		query     int     false      "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        search  		query     string  false      "Search value"
// @Success      200    		{object}  object{success=bool,message=string,data=[]models.UserResponse,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved user list"
// @Failure      400    		{object}  lib.ResponseError  "Invalid pagination parameters or page out of range"
// @Failure      500    		{object}  lib.ResponseError  "Internal server error while fetching or processing user data"
// @Router       /admin/users [get]
func ListUsers(ctx *gin.Context) {
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
			`SELECT COUNT(*) FROM users 
			LEFT JOIN profiles ON users.id = profiles.user_id 
			WHERE users.first_name ILIKE $1 
			OR users.last_name ILIKE $1
			OR profiles.phone_number ILIKE $1
			OR profiles.address ILIKE $1
			OR users.email ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&totalData)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total users in database",
			Error:   err.Error(),
		})
		return
	}

	offset := (page - 1) * limit
	var rows pgx.Rows
	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				users.id,
				COALESCE(profiles.image, '') AS image,
				COALESCE(users.first_name, '') AS first_name,
				COALESCE(users.last_name, '') AS last_name,
				COALESCE(profiles.phone_number, '') AS phone_number,
				COALESCE(profiles.address, '') AS address,
				users.email,
				users.role
			FROM users
			LEFT JOIN profiles ON users.id = profiles.user_id
			WHERE users.first_name ILIKE $3
			   OR users.last_name ILIKE $3
			   OR profiles.phone_number ILIKE $3
			   OR profiles.address ILIKE $3
			   OR users.email ILIKE $3
			ORDER BY users.id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(
			context.Background(),
			`SELECT 
				users.id,
				COALESCE(profiles.image, '') AS image,
				COALESCE(users.first_name, '') AS first_name,
				COALESCE(users.last_name, '') AS last_name,
				COALESCE(profiles.phone_number, '') AS phone_number,
				COALESCE(profiles.address, '') AS address,
				users.email,
				users.role
			FROM users
			LEFT JOIN profiles ON users.id = profiles.user_id
			ORDER BY users.id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch users from database",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.UserResponse])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process user data from database",
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
	baseURL := fmt.Sprintf("%s://%s/admin/users", scheme, host)

	var next any
	var prev any
	switch {
	case totalData == 0:
		page = 0
		next, prev = nil, nil
	case page == 1 && totalPage > 1:
		next = fmt.Sprintf("%s?page=%v&limit=%v", baseURL, page+1, limit)
		prev = nil
	case page == totalPage && totalPage > 1:
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
		"message": "Success get all user",
		"data":    users,
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

// DetailUser    godoc
// @Summary      Get detail user
// @Description  Retrieving detail user based on Id
// @Tags         admin/users
// @Accept 		 x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id   			path    int     true  "User Id"
// @Success      200  {object}  lib.ResponseSuccess{data=models.UserResponse}  "Successfully retrieved user"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "User not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while fetching users from database"
// @Router       /admin/users/{id} [get]
func DetailUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	user, message, err := models.GetUserById(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: message,
		Data:    user,
	})
}

// CreateUser    godoc
// @Summary      Create new user
// @Description  Create a new user with a unique email
// @Tags         admin/users
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        first_name     formData  string  true  "User first name"
// @Param        last_name      formData  string  true  "User last name"
// @Param        email          formData  string  true  "User email"
// @Param        password       formData  string  true  "User password"  format(password)
// @Param        phone          formData  string  false "User phone"
// @Param        address        formData  string  false "User address"
// @Param        role           formData  string  false "User role"  default(customer)
// @Param        profilephoto   formData  file    false "Profile photo (JPEG/PNG, max 1MB)"
// @Success      201  {object}  lib.ResponseSuccess{data=models.UserResponse}  "User created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or failed to hash password"
// @Failure      409  {object}  lib.ResponseError  "Email already registered"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while creating user"
// @Router       /admin/users [post]
func CreateUser(ctx *gin.Context) {
	var bodyCreateUser models.User
	err := ctx.ShouldBindWith(&bodyCreateUser, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	if bodyCreateUser.Role == "" {
		bodyCreateUser.Role = "customer"
	}

	var exists bool
	err = config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", bodyCreateUser.Email,
	).Scan(&exists)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking email uniqueness",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Email already registered",
		})
		return
	}

	hashedPassword, err := lib.HashPassword(bodyCreateUser.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}
	bodyCreateUser.Password = string(hashedPassword)

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
		`INSERT INTO users (first_name, last_name, email, role, password, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		bodyCreateUser.FirstName,
		bodyCreateUser.LastName,
		bodyCreateUser.Email,
		bodyCreateUser.Role,
		bodyCreateUser.Password,
		userIdFromToken,
		userIdFromToken,
	).Scan(&bodyCreateUser.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new user",
			Error:   err.Error(),
		})
		return
	}

	var savedFilePath string
	file, err := ctx.FormFile("profilephoto")
	if err == nil {
		if file.Size > 1<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "File size must be less than 1MB",
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
		fileName := fmt.Sprintf("user_%d_%d%s", bodyCreateUser.Id, time.Now().Unix(), ext)
		savedFilePath = "uploads/profiles/" + fileName

		if err := ctx.SaveUploadedFile(file, savedFilePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to save uploaded file",
				Error:   err.Error(),
			})
			return
		}
	}

	_, err = config.DB.Exec(
		context.Background(),
		`INSERT INTO profiles (user_id, image, address, phone_number, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		bodyCreateUser.Id,
		savedFilePath,
		bodyCreateUser.Address,
		bodyCreateUser.Phone,
		userIdFromToken,
		userIdFromToken,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new user profile",
			Error:   err.Error(),
		})
		return
	}

	profilePhoto := savedFilePath
	phone := bodyCreateUser.Phone
	address := bodyCreateUser.Address

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: "User created successfully",
		Data: models.UserResponse{
			Id:           bodyCreateUser.Id,
			ProfilePhoto: profilePhoto,
			FirstName:    bodyCreateUser.FirstName,
			LastName:     bodyCreateUser.LastName,
			Phone:        phone,
			Address:      address,
			Email:        bodyCreateUser.Email,
			Role:         bodyCreateUser.Role,
		},
	})
}

// UpdateUser    godoc
// @Summary      Update user
// @Description  Updating user data based on Id
// @Tags         admin/users
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path      int     true  "User Id"
// @Param        first_name     formData  string  false "User first name"
// @Param        last_name      formData  string  false "User last name"
// @Param        email          formData  string  false "User email"
// @Param        phone          formData  string  false "User phone"
// @Param        address        formData  string  false "User address"
// @Param        role           formData  string  false "User role"
// @Param        profilephoto   formData  file    false "Profile photo (JPEG/PNG, max 1MB)"
// @Success      200  {object}  lib.ResponseSuccess "User updated successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format or invalid request body"
// @Failure      404  {object}  lib.ResponseError  "User not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while updating user data"
// @Router       /admin/users/{id} [patch]
func UpdateUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	var bodyUpdate models.UserUpdateRequest
	err = ctx.ShouldBindWith(&bodyUpdate, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	if bodyUpdate.Role == "" {
		bodyUpdate.Role = "customer"
	}

	userIdFromToken, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	var savedFilePath string
	file, err := ctx.FormFile("profilephoto")
	if err == nil {
		if file.Size > 1<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "File size must be less than 1MB",
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
		fileName := fmt.Sprintf("user_%d_%d%s", id, time.Now().Unix(), ext)
		savedFilePath = "uploads/profiles/" + fileName

		if err := ctx.SaveUploadedFile(file, savedFilePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to save uploaded file",
				Error:   err.Error(),
			})
			return
		}
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE users 
		 SET first_name = COALESCE(NULLIF($1, ''), first_name),
		     last_name  = COALESCE(NULLIF($2, ''), last_name),
		     email      = COALESCE(NULLIF($3, ''), email),
		     role       = COALESCE(NULLIF($4, ''), role),
		     updated_by = $5,
		     updated_at = NOW()
		 WHERE id = $6`,
		bodyUpdate.FirstName,
		bodyUpdate.LastName,
		bodyUpdate.Email,
		bodyUpdate.Role,
		userIdFromToken,
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating user table",
			Error:   err.Error(),
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE profiles 
		 SET image        = COALESCE(NULLIF($1, ''), image),
		     address      = COALESCE(NULLIF($2, ''), address),
		     phone_number = COALESCE(NULLIF($3, ''), phone_number),
		     updated_by   = $4,
		     updated_at   = NOW()
		 WHERE user_id = $5`,
		savedFilePath,
		bodyUpdate.Address,
		bodyUpdate.Phone,
		userIdFromToken,
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating user profile",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "User updated successfully",
	})
}

// DeleteUser    godoc
// @Summary      Delete user
// @Description  Delete user by Id
// @Tags         admin/users
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Param        id             path    int     true  "User Id"
// @Success      200  {object}  lib.ResponseSuccess  "User deleted successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid Id format"
// @Failure      404  {object}  lib.ResponseError  "User not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while deleting user data"
// @Router       /admin/users/{id} [delete]
func DeleteUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid Id format",
			Error:   err.Error(),
		})
		return
	}

	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while deleting user data",
			Error:   err.Error(),
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "User not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "User deleted successfully",
	})
}
