package models

import "mime/multipart"

type Product struct {
	Id                int      `db:"id"`
	Images            []string `db:"images"`
	Name              string   `db:"name"`
	Description       string   `db:"description"`
	Price             float64  `db:"price"`
	DiscountPercent   float64  `db:"discount_percent"`
	Rating            float64  `db:"rating"`
	IsFlashSale       bool     `db:"is_flash_sale"`
	Stock             int      `db:"stock"`
	IsActive          bool     `db:"is_active"`
	SizeProducts      []string `db:"size_products"`
	ProductCategories []string `db:"product_categories"`
}

type ProductCreateRequest struct {
	Id                int
	Image1            *multipart.FileHeader `form:"image1"`
	Image2            *multipart.FileHeader `form:"image2"`
	Image3            *multipart.FileHeader `form:"image3"`
	Image4            *multipart.FileHeader `form:"image4"`
	Images            []string
	Name              string  `form:"name"`
	Description       string  `form:"description"`
	Price             float64 `form:"price"`
	DiscountPercent   float64 `form:"discount_percent"`
	Rating            float64 `form:"rating"`
	IsFlashSale       bool    `form:"is_flash_sale"`
	Stock             int     `form:"stock"`
	IsActive          bool    `form:"is_active"`
	SizeProducts      string  `form:"size_products"`
	ProductCategories string  `form:"product_categories"`
}
