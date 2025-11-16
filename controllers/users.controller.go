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

// ListUsers     godoc
// @Summary      Get list users
// @Description  Retrieving list users with pagination support and searching
// @Tags         admin/users
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header 	  string  true       "Bearer token"              default(Bearer <token>)
// @Param        page   		query     int     false      "Page number"  			 default(1)   minimum(1)
// @Param        limit  		query     int     false      "Number of items per page"  default(10)  minimum(1)  maximum(100)
// @Param        search  		query     string  false      "Search value"
// @Success      200    		{object}  object{success=bool,message=string,data=[]models.User,meta=object{currentPage=int,perPage=int,totalData=int,totalPages=int,next=string,prev=string}}  "Successfully retrieved user list"
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

	// get total data user
	totalData, err := models.GetTotalDataUsers(search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to count total users in database",
			Error:   err.Error(),
		})
		return
	}

	// get list all user
	users, message, err := models.GetListAllUser(page, limit, search)
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
	baseURL := fmt.Sprintf("%s://%s/admin/users", scheme, host)

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
// @Success      200  {object}  lib.ResponseSuccess{data=models.User}  "Successfully retrieved user"
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

	// get detail user
	user, message, err := models.GetDetailUser(id)
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
// @Param        fullName       formData  string  true  "User full name"
// @Param        email          formData  string  true  "User email"
// @Param        password       formData  string  true  "User password"  format(password)
// @Param        phone          formData  string  true  "User phone"
// @Param        address        formData  string  true  "User address"
// @Param        role           formData  string  true  "User role"  default(customer)
// @Param        filePhoto      formData  file    false  "Profile photo (JPEG/PNG, max 3MB)"
// @Success      201  {object}  lib.ResponseSuccess{data=models.User}  "User created successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or failed to hash password"
// @Failure      401  {object}  lib.ResponseError  "User unauthorized"
// @Failure      409  {object}  lib.ResponseError  "Email already registered"
// @Failure      500  {object}  lib.ResponseError  "Internal server error while creating user"
// @Router       /admin/users [post]
func CreateUser(ctx *gin.Context) {
	var bodyCreate models.User
	err := ctx.ShouldBindWith(&bodyCreate, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	if bodyCreate.Role == "" {
		bodyCreate.Role = "customer"
	}

	// check user email
	exists, err := models.CheckUserEmail(bodyCreate.Email)
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

	// hash password
	hashedPassword, err := lib.HashPassword(bodyCreate.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}
	bodyCreate.Password = string(hashedPassword)

	// get user id from token
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// get file from body request
	savedFilePath := ""
	file, err := ctx.FormFile("filePhoto")
	if err == nil {

		if file.Size > 3<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "File size must be less than 3MB",
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

		fileName := fmt.Sprintf("user_%d_%d%s", userId, time.Now().Unix(), ext)
		savedFilePath = "uploads/profiles/" + fileName

		os.MkdirAll("uploads/profiles", 0755)

		err = ctx.SaveUploadedFile(file, savedFilePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to save profile photo",
				Error:   err.Error(),
			})
			return
		}
		bodyCreate.ProfilePhoto = savedFilePath
	}

	// insert data user
	isSuccess, message, err := models.InsertDataUser(userId.(int), &bodyCreate, savedFilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: isSuccess,
		Message: "User created successfully",
		Data:    bodyCreate,
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
// @Param        fullName       formData  string  false "User full name"
// @Param        email          formData  string  false "User email"
// @Param        phone          formData  string  false "User phone"
// @Param        address        formData  string  false "User address"
// @Param        role           formData  string  false "User role"
// @Param        filePhoto      formData  file    false "Profile photo (JPEG/PNG, max 3MB)"
// @Success      200  {object}  lib.ResponseSuccess "User updated successfully"
// @Failure      400  {object}  lib.ResponseError   "Invalid Id format or invalid request body"
// @Failure      401  {object}  lib.ResponseError  "User unauthorized"
// @Failure      404  {object}  lib.ResponseError   "User not found"
// @Failure      500  {object}  lib.ResponseError   "Internal server error while updating user data"
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

	var bodyUpdate models.User
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
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	savedFilePath := ""
	file, err := ctx.FormFile("filePhoto")
	if err == nil {
		if file.Size > 3<<20 {
			ctx.JSON(http.StatusBadRequest, lib.ResponseError{
				Success: false,
				Message: "File size must be less than 3MB",
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

		fileName := fmt.Sprintf("user_%d_%d%s", userId, time.Now().Unix(), ext)
		savedFilePath = "uploads/profiles/" + fileName

		os.MkdirAll("uploads/profiles", 0755)

		err = ctx.SaveUploadedFile(file, savedFilePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "User updated but failed to save profile photo",
				Error:   err.Error(),
			})
			return
		}
		bodyUpdate.ProfilePhoto = savedFilePath
	}

	// insert data user
	isSuccess, message, err := models.UpdateDataUser(id, userId.(int), &bodyUpdate, savedFilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: isSuccess,
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

	commandTag, err := models.DeleteDataUser(id)
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
