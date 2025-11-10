package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

// VerifyResetToken    godoc
// @Summary      Verify password reset token
// @Description  Verify if the reset token is valid and not expired
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email  formData  string  true  "User email"
// @Param        token  formData  string  true  "6-digit reset token"
// @Success      200  {object}  lib.ResponseSuccess  "Token is valid"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or token format"
// @Failure      404  {object}  lib.ResponseError  "Invalid or expired token"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/verify-reset-token [post]
func VerifyResetToken(ctx *gin.Context) {
	var bodyRequest models.VerifyResetTokenRequest
	err := ctx.ShouldBindWith(&bodyRequest, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	var userId int
	var expiredAt time.Time
	err = config.DB.QueryRow(
		context.Background(),
		`SELECT pr.user_id, pr.expired_at
		 FROM password_resets pr
		 JOIN users u ON pr.user_id = u.id
		 WHERE u.email = $1 AND pr.token_reset = $2`,
		bodyRequest.Email,
		bodyRequest.Token,
	).Scan(&userId, &expiredAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Invalid token",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while verifying token",
			Error:   err.Error(),
		})
		return
	}

	if time.Now().After(expiredAt) {
		config.DB.Exec(
			context.Background(),
			"DELETE FROM password_resets WHERE user_id = $1",
			userId,
		)

		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Token has expired",
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Token is valid",
		Data: gin.H{
			"userId": userId,
		},
	})
}

// ResetPassword    godoc
// @Summary      Reset password
// @Description  Reset user password using valid token
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email        formData  string  true  "User email"
// @Param        token        formData  string  true  "6-digit reset token"
// @Param        new_password formData  string  true  "New password"  format(password)
// @Success      200  {object}  lib.ResponseSuccess  "Password reset successfully"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or failed to hash password"
// @Failure      404  {object}  lib.ResponseError  "Invalid or expired token"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/reset-password [patch]
func ResetPassword(ctx *gin.Context) {
	var bodyRequest models.ResetPasswordRequest
	err := ctx.ShouldBindWith(&bodyRequest, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	var userId int
	var expiredAt time.Time
	err = config.DB.QueryRow(
		context.Background(),
		`SELECT pr.user_id, pr.expired_at
		 FROM password_resets pr
		 JOIN users u ON pr.user_id = u.id
		 WHERE u.email = $1 AND pr.token_reset = $2`,
		bodyRequest.Email,
		bodyRequest.Token,
	).Scan(&userId, &expiredAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "Invalid token",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while verifying token",
			Error:   err.Error(),
		})
		return
	}

	if time.Now().After(expiredAt) {
		config.DB.Exec(
			context.Background(),
			"DELETE FROM password_resets WHERE user_id = $1",
			userId,
		)

		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Token has expired",
		})
		return
	}

	hashedPassword, err := lib.HashPassword(bodyRequest.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		`UPDATE users 
		 SET password = $1, 
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		hashedPassword,
		userId,
		userId,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while updating password",
			Error:   err.Error(),
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		"DELETE FROM password_resets WHERE user_id = $1",
		userId,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while cleaning token",
			Error:   err.Error(),
		})
		return
	}

	_, err = config.DB.Exec(
		context.Background(),
		"UPDATE sessions SET is_active = false WHERE user_id = $1",
		userId,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while non-active session user",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Password has been reset successfully",
	})
}
