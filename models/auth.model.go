package models

type Register struct {
	Id        int    `json:"id"`
	FirstName string `form:"firstName" json:"firstName"`
	LastName  string `form:"lastName" json:"lastName"`
	Email     string `form:"email" json:"email"`
	Password  string `form:"password" json:"-"`
	Role      string `form:"role" json:"role"`
}
