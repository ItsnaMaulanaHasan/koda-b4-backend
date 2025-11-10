package models

import "time"

type Category struct {
	Id        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	CreatedBy int       `db:"created_by" json:"created_by"`
	UpdatedBy int       `db:"updated_by" json:"updated_by"`
}

type CategoryResponse struct {
	Id        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
