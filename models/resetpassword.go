package models

import "time"

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
