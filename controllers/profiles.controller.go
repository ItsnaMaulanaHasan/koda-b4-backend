package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"backend-daily-greens/utils"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// DetailProfile    godoc
// @Summary         Get detail profile
// @Description     Retrieving detail profile based on Id in token
// @Tags            profiles
// @Produce         json
// @Security        BearerAuth
// @Param           Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Success         200  {object}  lib.ResponseSuccess{data=models.User}  "Successfully retrieved user"
// @Failure         401  {object}  lib.ResponseError  "User Id not found in token"
// @Failure         404  {object}  lib.ResponseError  "User not found"
// @Failure         500  {object}  lib.ResponseError  "Internal server error while fetching profiles from database"
// @Router          /profiles [get]
func DetailProfile(ctx *gin.Context) {
	// get user id from token
	userId, exist := ctx.Get("userId")
	if !exist {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	// get detail profile
	user, message, err := models.GetDetailProfile(userId.(int))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "User not found" {
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
		Message: "Success get data profile",
		Data:    user,
	})
}

// UpdateProfiles    godoc
// @Summary          Update profile
// @Description      Updating user profile based on Id from token
// @Tags             profiles
// @Accept           x-www-form-urlencoded
// @Produce          json
// @Security         BearerAuth
// @Param            Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param            fullName      formData  string  false "User fullname name"
// @Param            email          formData  string  false "User email"
// @Param            phone          formData  string  false "User phone"
// @Param            address        formData  string  false "User address"
// @Success          200  {object}  lib.ResponseSuccess "User updated successfully"
// @Failure          400  {object}  lib.ResponseError  "Invalid request body"
// @Failure          404  {object}  lib.ResponseError  "User not found"
// @Failure          500  {object}  lib.ResponseError  "Internal server error while updating user data"
// @Router           /profiles [patch]
func UpdateProfile(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	var bodyUpdate models.ProfileRequest
	err := ctx.ShouldBindWith(&bodyUpdate, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	isSuccess, message, dataUpdated, err := models.UpdateDataProfile(userId.(int), bodyUpdate)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "User not found" || message == "Profile not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: isSuccess,
		Message: message,
		Data:    dataUpdated,
	})
}

// UploadProfilePhoto  godoc
// @Summary            Upload photo profile user
// @Description        Upload photo profie user data based on Id from token
// @Tags               profiles
// @Accept             x-www-form-urlencoded
// @Produce            json
// @Security           BearerAuth
// @Param              Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Param              profilePhoto   formData  file    false "Profile photo (JPEG/PNG, max 3MB)"
// @Success            200  {object}  lib.ResponseSuccess "User updated successfully"
// @Failure            400  {object}  lib.ResponseError  "Invalid request body"
// @Failure            404  {object}  lib.ResponseError  "User not found"
// @Failure            401  {object}  lib.ResponseError  "User Id not found in token"
// @Failure            500  {object}  lib.ResponseError  "Internal server error while upload profile photo"
// @Router             /profiles/photo [patch]
func UploadProfilePhoto(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	file, err := ctx.FormFile("profilePhoto")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "File is required",
			Error:   err.Error(),
		})
		return
	}

	if file.Size > 3<<20 {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "File size must be less than 3MB",
		})
		return
	}

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

	// check content type
	contentType := file.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Image has invalid type. Only JPEG and PNG are allowed",
		})
		return
	}

	// check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Image has invalid extension. Only JPG and PNG are allowed",
		})
		return
	}

	fileName := fmt.Sprintf("user_%d_%d", userId, time.Now().Unix())
	imageUrl, err := utils.UploadToSupabase(file, fileName, "profile-photos")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to upload profile photo",
			Error:   err.Error(),
		})
		return
	}

	isSuccess, message, err := models.UploadProfilePhotoUser(userId.(int), imageUrl)
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
		Message: message,
		Data: gin.H{
			"profilePhoto": imageUrl,
		},
	})
}
