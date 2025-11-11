package models

import "time"

type Category struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" form:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy int       `json:"created_by,omitempty" db:"-"`
	UpdatedBy int       `json:"updated_by,omitempty" db:"-"`
}
