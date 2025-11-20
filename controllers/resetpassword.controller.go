package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetTokenReset  godoc
// @Summary       Request password reset
// @Description   Send a 12-digit reset token to user's email
// @Tags          auth
// @Accept        x-www-form-urlencoded
// @Produce       json
// @Param         email  formData  string  true  "User email"
// @Success       200  {object}  lib.ResponseSuccess "Reset token sent successfully"
// @Failure       400  {object}  lib.ResponseError   "Invalid request body"
// @Failure       404  {object}  lib.ResponseError   "Email not found"
// @Failure       500  {object}  lib.ResponseError   "Internal server error"
// @Router        /auth/forgot-password [post]
func GetTokenReset(ctx *gin.Context) {
	email := ctx.PostForm("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	// get user id
	userId, message, err := models.GetUserIdByEmail(email)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Email not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// generate token
	token := lib.GenerateRandomToken(12)

	// delete old tokens
	err = models.DeleteOldPasswordResetTokens(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while cleaning old tokens",
			Error:   err.Error(),
		})
		return
	}

	// insert token
	message, err = models.InsertPasswordResetToken(userId, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// build reset URL
	resetURL := fmt.Sprintf(
		"%s/reset-password?otp=%s&email=%s",
		os.Getenv("ORIGIN_URL"),
		token,
		email,
	)

	// email HTML content
	htmlContent := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>Click the link below to reset your password:</p>
		<p><a href="%s">%s</a></p>
		<br/>
		<p>If you did not request this, you can ignore this email.</p>
	`, resetURL, resetURL)

	// send via resend
	err = lib.SendResendEmail(email, "Password Reset Request", htmlContent)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to send reset email",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Password reset link sent to email",
		Data: gin.H{
			"email": email,
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

	// verify token
	userId, expiredAt, message, err := models.VerifyPasswordResetToken(bodyRequest.Email, bodyRequest.Token)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Invalid token" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// check if token expired
	if time.Now().After(expiredAt) {
		models.DeleteOldPasswordResetTokens(userId)

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
// @Param        newPassword formData  string  true  "New password"  format(password)
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

	// verify token
	userId, expiredAt, message, err := models.VerifyPasswordResetToken(bodyRequest.Email, bodyRequest.Token)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Invalid token" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// check if token expired
	if time.Now().After(expiredAt) {
		models.DeleteOldPasswordResetTokens(userId)

		ctx.JSON(http.StatusNotFound, lib.ResponseError{
			Success: false,
			Message: "Token has expired",
		})
		return
	}

	// hash new password
	hashedPassword, err := lib.HashPassword(bodyRequest.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}

	// update user password
	isSuccess, message, err := models.UpdateUserPassword(userId, string(hashedPassword))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: isSuccess,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	// delete used token
	err = models.DeleteOldPasswordResetTokens(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Internal server error while cleaning token",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Password has been reset successfully",
	})
}
