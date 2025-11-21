package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserProfile struct {
	Id           int       `db:"id" json:"id"`
	ProfilePhoto string    `db:"profile_photo" json:"profilePhoto"`
	FullName     string    `db:"full_name" json:"fullName"`
	Email        string    `db:"email" json:"email"`
	Phone        string    `db:"phone_number" json:"phone"`
	Address      string    `db:"address" json:"address"`
	Role         string    `db:"role" json:"role"`
	JoinDate     time.Time `db:"created_at" json:"joinDate"`
}

type ProfileRequest struct {
	FullName *string `form:"fullName" json:"fullName,omitempty"`
	Email    *string `form:"email" json:"email,omitempty"`
	Phone    *string `form:"phone" json:"phone,omitempty"`
	Address  *string `form:"address" json:"address,omitempty"`
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
			COALESCE(p.phone_number, '') AS phone_number,
			u.role,
			u.created_at
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

func UpdateDataProfile(userId int, bodyUpdate ProfileRequest) (bool, string, UserProfile, error) {
	isSuccess := false
	message := ""
	var userProfile UserProfile

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, userProfile, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		`UPDATE users 
         SET email      = COALESCE(NULLIF($1, ''), email),
             updated_by = $2,
             updated_at = NOW()
         WHERE id = $2
         RETURNING id, email, role, created_at`,
		bodyUpdate.Email,
		userId,
	).Scan(&userProfile.Id, &userProfile.Email, &userProfile.Role, &userProfile.JoinDate)

	if err != nil {
		if err == pgx.ErrNoRows {
			message = "User not found"
			return isSuccess, message, userProfile, nil
		}
		message = "Internal server error while updating user table"
		return isSuccess, message, userProfile, err
	}

	err = tx.QueryRow(
		ctx,
		`UPDATE profiles 
		 SET full_name    = COALESCE(NULLIF($1, ''), full_name),
		     address      = COALESCE(NULLIF($2, ''), address),
		     phone_number = COALESCE(NULLIF($3, ''), phone_number),
		     updated_by   = $4,
		     updated_at   = NOW()
		 WHERE user_id = $4
		 RETURNING COALESCE(profile_photo, '') AS profile_photo, full_name, COALESCE(address, '') AS address, COALESCE(phone_number, '') AS phone_number`,
		bodyUpdate.FullName,
		bodyUpdate.Address,
		bodyUpdate.Phone,
		userId,
	).Scan(
		&userProfile.ProfilePhoto,
		&userProfile.FullName,
		&userProfile.Address,
		&userProfile.Phone,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			message = "Profile not found"
			return isSuccess, message, userProfile, nil
		}
		message = "Internal server error while updating user profile"
		return isSuccess, message, userProfile, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, userProfile, err
	}

	isSuccess = true
	message = "User updated successfully"
	return isSuccess, message, userProfile, nil
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
		 WHERE user_id = $2`,
		savedFilePath,
		userId,
	)
	if err != nil {
		message = "Internal server error while updating user profile"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "User updated successfully"
	return isSuccess, message, nil
}
