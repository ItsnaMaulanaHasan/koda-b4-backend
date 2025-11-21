package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matthewhartstonge/argon2"
)

// Register      godoc
// @Summary      Register new user
// @Description  Create a new user with a unique email
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        fullName  formData  string  true  "Full name user"
// @Param        email     formData  string  true  "Email user"
// @Param        password  formData  string  true  "Password user" format(password)
// @Param        role      formData  string  true  "Role user" default(customer)
// @Success      201  {object}  lib.ResponseSuccess{data=models.Register}  "User created successfully."
// @Failure      400  {object}  lib.ResponseError  "Invalid request body or failed to hash password."
// @Failure      409  {object}  lib.ResponseError  "Email already registered."
// @Failure      500  {object}  lib.ResponseError  "Internal server error while creating user."
// @Router       /auth/register [post]
func Register(ctx *gin.Context) {
	var bodyRegister models.Register
	err := ctx.ShouldBindWith(&bodyRegister, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Please provide valid registration information",
			Error:   err.Error(),
		})
		return
	}

	if bodyRegister.Role == "" {
		bodyRegister.Role = "customer"
	}

	// check user email
	exists, err := models.CheckUserEmail(bodyRegister.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Unable to process registration. Please try again",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, lib.ResponseError{
			Success: false,
			Message: "Email is already registered",
		})
		return
	}

	// hash password
	hashedPassword, err := lib.HashPassword(bodyRegister.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}
	bodyRegister.Password = string(hashedPassword)

	isSuccess, message, err := models.RegisterUser(&bodyRegister)
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
		Message: message,
		Data: models.Register{
			Id:       bodyRegister.Id,
			FullName: bodyRegister.FullName,
			Email:    bodyRegister.Email,
			Role:     bodyRegister.Role,
		},
	})
}

// Login         godoc
// @Summary      Login user
// @Description  Log in with existing email data
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email     formData  string  true  "User email"
// @Param        password  formData  string  true  "User password" format(password)
// @Success      200  {object}  lib.ResponseSuccess{data=object{token=string}}  "Login successful"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      401  {object}  lib.ResponseError  "Incorrect email or password"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/login [post]
func Login(ctx *gin.Context) {
	var bodyLogin models.Login
	err := ctx.ShouldBindWith(&bodyLogin, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Please provide valid email and password",
			Error:   err.Error(),
		})
		return
	}

	user, message, err := models.GetUserByEmail(&bodyLogin)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if message == "Incorrect email or password" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, lib.ResponseError{
			Success: false,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	isPasswordValid, err := argon2.VerifyEncoded(
		[]byte(bodyLogin.Password),
		[]byte(user.Password),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to verifi password. Please try again",
			Error:   err.Error(),
		})
		return
	}

	if !isPasswordValid {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Incorrect email or password",
		})
		return
	}

	jwtToken, err := lib.GenerateToken(user.Id, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to generate token. Please try again",
			Error:   err.Error(),
		})
		return
	}

	userTokenKey := fmt.Sprintf("user_%d_token", user.Id)

	err = config.Rdb.Set(context.Background(), userTokenKey, jwtToken, 24*time.Hour).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to save token. Please try again",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "Login successful!",
		Data: gin.H{
			"token": jwtToken,
		},
	})
}

// Logout        godoc
// @Summary      Logout user
// @Description  Logout user by blacklisting the JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer token"  default(Bearer <token>)
// @Success      200  {object}  lib.ResponseSuccess  "User logged out successfully"
// @Failure      401  {object}  lib.ResponseError    "User unauthorized or token invalid"
// @Failure      500  {object}  lib.ResponseError    "Internal server error"
// @Router       /auth/logout [post]
func Logout(ctx *gin.Context) {
	authHeader := ctx.Request.Header.Get("Authorization")
	tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
	if !found {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Authorization header required or invalid format",
		})
		ctx.Abort()
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &lib.UserPayload{}, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("APP_SECRET")), nil
	})

	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Invalid or expired token",
			Error:   err.Error(),
		})
		return
	}

	claims, ok := token.Claims.(*lib.UserPayload)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Invalid token claims",
		})
		return
	}

	expiryTime := claims.ExpiresAt.Time
	ttl := time.Until(expiryTime)

	if ttl <= 0 {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Token already expired",
		})
		return
	}

	blacklistKey := "blacklist:" + tokenString

	err = config.Rdb.Set(context.Background(), blacklistKey, tokenString, ttl).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to logout user",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "User logged out successfully",
	})
}
