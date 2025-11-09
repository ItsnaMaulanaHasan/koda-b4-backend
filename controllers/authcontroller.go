package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Register godoc
// @Summary      Create new user
// @Description  Create a new user with a unique email
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        user formData  models.User true "User registration data"
// @Success      201  {object}  lib.ResponseSuccess{data=models.UserResponse}  "User created successfully."
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or failed to hash password."
// @Failure      409  {object}  lib.ResponseError  "Email already registered."
// @Failure      500  {object}  lib.ResponseError  "Internal server error while creating user."
// @Router       /auth/register [post]
func Register(ctx *gin.Context) {
	var bodyCreateUser models.User
	err := ctx.ShouldBindWith(&bodyCreateUser, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: true,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	var exists bool
	err = config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", bodyCreateUser.Email,
	).Scan(&exists)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking email uniqueness.",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Email already registered.",
		})
		return
	}

	hashedPassword, err := lib.HashPassword(bodyCreateUser.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password.",
			Error:   err.Error(),
		})
		return
	}
	bodyCreateUser.Password = string(hashedPassword)

	err = config.DB.QueryRow(
		context.Background(),
		`INSERT INTO users (first_name, last_name, email, role, password)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		bodyCreateUser.FirstName, bodyCreateUser.LastName, bodyCreateUser.Email, bodyCreateUser.Role, bodyCreateUser.Password,
	).Scan(&bodyCreateUser.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while inserting new user.",
			Error:   err.Error(),
		})
		return
	}

	models.ResponseUserData = &models.UserResponse{
		Id:        bodyCreateUser.Id,
		FirstName: bodyCreateUser.FirstName,
		LastName:  bodyCreateUser.LastName,
		Email:     bodyCreateUser.Email,
		Role:      bodyCreateUser.Role,
	}

	ctx.JSON(http.StatusCreated, lib.ResponseSuccess{
		Success: true,
		Message: "User created successfully.",
		Data:    models.ResponseUserData,
	})
}
