package models

type Product struct {
	Id                int      `json:"id" db:"id"`
	Image             []string `json:"image" db:"image" form:"image"`
	Name              string   `json:"name" db:"name" form:"name"`
	Description       string   `json:"description" db:"description" form:"description"`
	Price             float64  `json:"price" db:"price" form:"price"`
	DiscountPercent   float64  `json:"discount_percent" db:"discount_percent" form:"discount_percent"`
	Rating            float64  `json:"rating" db:"rating" form:"rating"`
	IsFlashSale       bool     `json:"is_flash_sale" db:"is_flash_sale" form:"is_flash_sale"`
	Stock             int      `json:"stock" db:"stock" form:"stock"`
	IsActive          bool     `json:"is_active" db:"is_active" form:"is_active"`
	SizeProducts      []string `json:"size_products" db:"size_products" form:"size_products"`
	ProductCategories []string `json:"product_categories" db:"product_categories" form:"product_categories"`
}
