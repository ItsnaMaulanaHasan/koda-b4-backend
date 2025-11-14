package models

type Register struct {
	Id       int    `json:"id"`
	FullName string `form:"fullName" json:"fullName"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"-"`
	Role     string `form:"role" json:"role"`
}
