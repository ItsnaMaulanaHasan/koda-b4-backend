package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// DetailProfile    godoc
// @Summary         Get detail profile
// @Description     Retrieving detail profile based on Id in token
// @Tags            profiles
// @Produce         json
// @Security        BearerAuth
// @Param           Authorization  header  string  true  "Bearer token"  default(Bearer <token>)
// @Success         200  {object}  lib.ResponseSuccess{data=models.User}  "Successfully retrieved user"
// @Failure         401  {object}  lib.ResponseError  "User unathorized"
// @Failure         404  {object}  lib.ResponseError  "User not found"
// @Failure         500  {object}  lib.ResponseError  "Internal server error while fetching profiles from database"
// @Router          /profiles [get]
func DetailProfile(ctx *gin.Context) {
	userId, exist := ctx.Get("userId")
	if !exist {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "User Id not found in token",
		})
		return
	}

	user, message, err := models.GetProfileById(userId.(int))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: message,
			})
		}
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Success get data profile",
		Data:    user,
	})
}

func UpdateProfile(ctx *gin.Context) {

}

func UploadPhotoProfile(ctx *gin.Context) {

}
