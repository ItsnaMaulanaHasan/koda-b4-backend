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
	NewPassword string `form:"newPassword" json:"newPassword" binding:"required,min=6"`
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

func InsertPasswordResetToken(userId int, token string) (string, error) {
	message := ""
	_, err := config.DB.Exec(
		context.Background(),
		`INSERT INTO password_resets (user_id, token_reset, created_by, updated_by)
		 VALUES ($1, $2, $3, $4)`,
		userId,
		token,
		userId,
		userId,
	)

	if err != nil {
		message = "Internal server error while creating reset token"
		return message, err
	}

	message = "Reset token created successfully"
	return message, nil
}

func VerifyPasswordResetToken(email string, token string) (int, time.Time, string, error) {
	var userId int
	var expiredAt time.Time
	message := ""

	err := config.DB.QueryRow(
		context.Background(),
		`SELECT pr.user_id, pr.expired_at
		 FROM password_resets pr
		 JOIN users u ON pr.user_id = u.id
		 WHERE u.email = $1 AND pr.token_reset = $2`,
		email,
		token,
	).Scan(&userId, &expiredAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Invalid token"
			return userId, expiredAt, message, err
		}
		message = "Internal server error while verifying token"
		return userId, expiredAt, message, err
	}

	message = "Token verified"
	return userId, expiredAt, message, nil
}

func UpdateUserPassword(userId int, hashedPassword string) (bool, string, error) {
	isSuccess := false
	message := ""

	_, err := config.DB.Exec(
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
		message = "Internal server error while updating password"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "Password updated successfully"
	return isSuccess, message, nil
}

func DeactivateAllUserSessions(userId int) error {
	_, err := config.DB.Exec(
		context.Background(),
		`UPDATE sessions 
		 SET is_active = false,
		     logout_time = NOW(),
		     updated_by = $1,
		     updated_at = NOW()
		 WHERE user_id = $1 AND is_active = true`,
		userId,
	)
	return err
}
