package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type UserProfile struct {
	Id           int    `db:"id" json:"id"`
	ProfilePhoto string `db:"profile_photo" json:"profilePhoto"`
	FullName     string `db:"full_name" json:"fullName"`
	Email        string `db:"email" json:"email"`
	Phone        string `db:"phone" json:"phone"`
	Address      string `db:"address" json:"address"`
	Password     string `db:"password" json:"-"`
}

type ProfileRequest struct {
	FullName string `form:"fullName" json:"fullName"`
	Email    string `form:"email" json:"email"`
	Phone    string `form:"phone" json:"phone"`
	Address  string `form:"address" json:"address"`
}

func GetDetailProfile(userId int) (UserProfile, string, error) {
	user := UserProfile{}
	message := ""

	// get detail profile
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			u.id,
			COALESCE(p.profile_photo, '') AS profile_photo,
			p.full_name,
			u.email,
			COALESCE(p.address, '') AS address,
			COALESCE(p.phone_number, '') AS phone,
			u.password
		FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE u.id = $1`, userId)
	if err != nil {
		message = "Internal server error while fetch data profile"
		return user, message, err
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[UserProfile])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "User not found"
			return user, message, err
		}
		message = "Internal server error while process data profile"
		return user, message, err
	}

	message = "Success get profile user"
	return user, message, nil
}

func UpdateDataProfile(userId int, bodyUpdate ProfileRequest) (bool, string, error) {
	isSuccess := false
	message := ""

	// start transaction database
	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	// update data user
	commandTag, err := tx.Exec(
		ctx,
		`UPDATE users 
		 SET email      = COALESCE(NULLIF($3, ''), email),
		     role       = COALESCE(NULLIF($4, ''), role),
		     updated_by = $5,
		     updated_at = NOW()
		 WHERE id = $6`,
		bodyUpdate.FullName,
		bodyUpdate.Email,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user table"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "User not found"
		return isSuccess, message, nil
	}

	// update data user profile
	commandTag, err = tx.Exec(
		ctx,
		`UPDATE profiles 
		 SET full_name    = COALESCE(NULLIF($1, ''), full_name),
		     address      = COALESCE(NULLIF($2, ''), address),
		     phone_number = COALESCE(NULLIF($3, ''), phone_number),
		     updated_by   = $4,
		     updated_at   = NOW()
		 WHERE user_id = $5`,
		bodyUpdate.FullName,
		bodyUpdate.Address,
		bodyUpdate.Phone,
		userId,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user profile"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Profile not found"
		return isSuccess, message, nil
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "User updated successfully"
	return isSuccess, message, nil
}

func UploadProfilePhotoUser(userId int, savedFilePath string) (bool, string, error) {
	isSuccess := false
	message := ""
	_, err := config.DB.Exec(
		context.Background(),
		`UPDATE profiles 
		 SET profile_photo = COALESCE($1, profile_photo),
		     updated_by    = $2,
		     updated_at    = NOW()
		 WHERE user_id = $3`,
		savedFilePath,
		userId,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user profile"
		return isSuccess, message, err
	}

	message = "User updated successfully"
	return isSuccess, message, nil
}
