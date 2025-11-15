package controllers

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
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
			Message: "Invalid form data",
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
	hashedPassword, err := lib.HashPassword(bodyRegister.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
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
// @Failure      401  {object}  lib.ResponseError  "Invalid email or password"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/login [post]
func Login(ctx *gin.Context) {
	var bodyLogin models.Login
	err := ctx.ShouldBindWith(&bodyLogin, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	user, message, err := models.GetUserByEmail(&bodyLogin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: message,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
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
			Message: "Failed to verify password",
			Error:   err.Error(),
		})
		return
	}

	if !isPasswordValid {
		ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	jwtToken, err := lib.GenerateToken(user.Id, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to generate token",
			Error:   err.Error(),
		})
		return
	}

	token, err := jwt.ParseWithClaims(jwtToken, &lib.UserPayload{}, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("APP_SECRET")), nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to parse token",
			Error:   err.Error(),
		})
		return
	}

	claims, ok := token.Claims.(*lib.UserPayload)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to extract token claims",
		})
		return
	}

	var session = models.Session{
		UserId:    user.Id,
		Token:     jwtToken,
		LoginTime: claims.IssuedAt.Time,
		ExpiredAt: claims.ExpiresAt.Time,
		IpAddress: ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
	}

	err = models.CreateSession(&session)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to create session",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, lib.ResponseSuccess{
		Success: true,
		Message: "User login successfully",
		Data: gin.H{
			"token":     session.Token,
			"sessionId": session.Id,
			"loginTime": session.LoginTime,
			"expiresAt": session.ExpiredAt,
		},
	})
}
