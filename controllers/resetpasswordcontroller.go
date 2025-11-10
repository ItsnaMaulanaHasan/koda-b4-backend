package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// ForgotPassword  godoc
// @Summary      Request password reset
// @Description  Send a 6-digit reset token to user's email
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email  formData  string  true  "User email"
// @Success      200  {object}  lib.ResponseSuccess  "Reset token sent successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      404  {object}  lib.ResponseError  "Email not found"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/forgot-password [post]
func GetTokenReset(ctx *gin.Context) {
	bodyRequest := ctx.PostForm("email")

	var userId int
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT id FROM users WHERE email = $1",
		bodyRequest,
	).Scan(&userId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Email not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while checking email",
			Error:   err.Error(),
		})
		return
	}

	token := fmt.Sprintf("%06d", rand.Intn(1000000))

	_, err = config.DB.Exec(
		context.Background(),
		"DELETE FROM password_resets WHERE user_id = $1",
		userId,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while cleaning old tokens",
			Error:   err.Error(),
		})
		return
	}

	var resetId int
	err = config.DB.QueryRow(
		context.Background(),
		`INSERT INTO password_resets (user_id, token_reset, created_by, updated_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		userId,
		token,
		userId,
		userId,
	).Scan(&resetId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while creating reset token",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Password reset token has been sent to your email",
		Data: gin.H{
			"email": bodyRequest,
			"token": token,
		},
	})
}
