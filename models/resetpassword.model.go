package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type PasswordReset struct {
	Id         int       `db:"id" json:"id"`
	UserId     int       `db:"user_id" json:"userId"`
	TokenReset string    `db:"token_reset" json:"tokenReset"`
	ExpiredAt  time.Time `db:"expired_at" json:"expiredAt"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time `db:"updated_at" json:"updatedAt"`
	CreatedBy  int       `db:"created_by" json:"createdBy"`
	UpdatedBy  int       `db:"updated_by" json:"updatedBy"`
}

type VerifyResetTokenRequest struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Token string `form:"token" json:"token" binding:"required,len=6"`
}

type ResetPasswordRequest struct {
	Email       string `form:"email" json:"email" binding:"required,email"`
	Token       string `form:"token" json:"token" binding:"required,len=6"`
	NewPassword string `form:"new_password" json:"newPassword" binding:"required,min=6"`
}

func GetUserIdByEmail(email string) (int, string, error) {
	var userId int
	message := ""

	err := config.DB.QueryRow(
		context.Background(),
		"SELECT id FROM users WHERE email = $1",
		email,
	).Scan(&userId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Email not found"
			return userId, message, err
		}
		message = "Internal server error while checking email"
		return userId, message, err
	}

	message = "User found"
	return userId, message, nil
}

func DeleteOldPasswordResetTokens(userId int) error {
	_, err := config.DB.Exec(
		context.Background(),
		"DELETE FROM password_resets WHERE user_id = $1",
		userId,
	)
	return err
}

func InsertPasswordResetToken(userId int, token string) (int, string, error) {
	var resetId int
	message := ""

	err := config.DB.QueryRow(
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
		message = "Internal server error while creating reset token"
		return resetId, message, err
	}

	message = "Reset token created successfully"
	return resetId, message, nil
}
