package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Register struct {
	Id       int    `json:"id"`
	FullName string `form:"fullName" json:"fullName"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"-"`
	Role     string `form:"role" json:"role"`
}

type Login struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

type QueryLogin struct {
	Id       int    `db:"id"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

type Session struct {
	Id         int `json:"sessionId"`
	UserId     int
	LoginTime  time.Time  `json:"loginTime"`
	LogoutTime *time.Time `json:"logoutTime,omitempty"`
	ExpiredAt  time.Time  `json:"expiredAt"`
	IpAddress  string
	UserAgent  string
	IsActive   bool
}

func RegisterUser(bodyRegister *Register) (bool, string, error) {
	isSuccess := false
	message := ""

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	// insert data to users
	err = tx.QueryRow(
		ctx,
		`INSERT INTO users (email, role, password)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		bodyRegister.Email, bodyRegister.Role, bodyRegister.Password,
	).Scan(&bodyRegister.Id)
	if err != nil {
		message = "Internal server error while inserting new user"
		return isSuccess, message, err
	}

	// update created_by and updated_by
	_, err = tx.Exec(ctx, `UPDATE users SET created_by = $1, updated_by = $1 WHERE id = $1`, bodyRegister.Id)
	if err != nil {
		message = "Internal server error while update created_by and updated_by"
		return isSuccess, message, err
	}

	// insert data to profiles
	_, err = tx.Exec(ctx, `INSERT INTO profiles (user_id, full_name, created_by, updated_by) VALUES ($1, $2, $3, $4)`,
		bodyRegister.Id, bodyRegister.FullName, bodyRegister.Id, bodyRegister.Id,
	)
	if err != nil {
		message = "Internal server error while inserting new profile"
		return isSuccess, message, err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "User registered successfully"
	return isSuccess, message, nil
}

func GetUserByEmail(bodyLogin *Login) (QueryLogin, string, error) {
	message := ""
	user := QueryLogin{}
	rows, err := config.DB.Query(context.Background(),
		"SELECT id, password, role FROM users WHERE email = $1",
		bodyLogin.Email,
	)
	if err != nil {
		message = "Failed to fetch user from database"
		return user, message, err
	}

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[QueryLogin])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "User not found"
			return user, message, err
		}

		message = "Failed to process user data"
		return user, message, err
	}

	return user, message, nil
}

func CreateSession(session *Session) error {
	err := config.DB.QueryRow(
		context.Background(),
		`INSERT INTO sessions 
		(user_id, login_time, expired_at, ip_address, device, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		session.UserId,
		session.LoginTime,
		session.ExpiredAt,
		session.IpAddress,
		session.UserAgent,
		true,
		session.UserId,
		session.UserId,
	).Scan(&session.Id)

	if err != nil {
		return err
	}

	return nil
}

func LogoutSession(userId int, sessionId int) (bool, string, error) {
	isSuccess := false
	message := ""

	commandTag, err := config.DB.Exec(
		context.Background(),
		`UPDATE sessions 
		 SET is_active = false,
		 	 logout_time = NOW(),
		     updated_by = $1, 
		     updated_at = NOW() 
		 WHERE id = $2 AND user_id = $1 AND is_active = true`,
		userId,
		sessionId,
	)
	if err != nil {
		message = "Internal server error while updating session"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Session not found or already logged out"
		return isSuccess, message, nil
	}

	isSuccess = true
	message = "User logged out successfully"
	return isSuccess, message, nil
}

func IsSessionActive(sessionId int) (bool, error) {
	var isActive bool
	err := config.DB.QueryRow(
		context.Background(),
		`SELECT is_active FROM sessions 
         WHERE id = $1 AND expired_at > NOW()`,
		sessionId,
	).Scan(&isActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return isActive, nil
}
