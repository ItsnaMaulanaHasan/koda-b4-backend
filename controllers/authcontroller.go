package controllers

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/matthewhartstonge/argon2"
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

// Login godoc
// @Summary      Login user
// @Description  Log in with existing email data
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email     formData  string  true  "User email"
// @Param        password  formData  string  true  "User password"
// @Success      200  {object}  lib.ResponseSuccess{data=object{token=string}}  "Login successful"
// @Failure      400  {object}  lib.ResponseError  "Invalid request body"
// @Failure      401  {object}  lib.ResponseError  "Invalid email or password"
// @Failure      500  {object}  lib.ResponseError  "Internal server error"
// @Router       /auth/login [post]
func Login(ctx *gin.Context) {
	var loginData struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=6"`
	}

	err := ctx.ShouldBindWith(&loginData, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, lib.ResponseError{
			Success: false,
			Message: "Invalid form data",
			Error:   err.Error(),
		})
		return
	}

	rows, err := config.DB.Query(context.Background(),
		"SELECT id, first_name, last_name, email, password, role FROM users WHERE email = $1",
		loginData.Email,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to fetch user from database.",
			Error:   err.Error(),
		})
		return
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, lib.ResponseError{
				Success: false,
				Message: "User not found.",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
			Success: false,
			Message: "Failed to process user data.",
			Error:   err.Error(),
		})
		return
	}

	isPasswordValid, err := argon2.VerifyEncoded(
		[]byte(loginData.Password),
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

	jwtToken, err := lib.GenerateToken(user.Id)
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

	loginTime := claims.IssuedAt.Time
	expiredAt := claims.ExpiresAt.Time
	ipAddress := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	var sessionId int
	err = config.DB.QueryRow(
		context.Background(),
		`INSERT INTO sessions 
		(user_id, session_token, login_time, expired_at, ip_address, device, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		user.Id,
		jwtToken,
		loginTime,
		expiredAt,
		ipAddress,
		userAgent,
		true,
		user.Id,
		user.Id,
	).Scan(&sessionId)

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
			"token":      jwtToken,
			"session_id": sessionId,
			"expires_at": expiredAt,
		},
	})
}
