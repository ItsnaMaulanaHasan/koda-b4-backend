package models

import (
	"backend-daily-greens/config"
	"context"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
)

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

type ProductRequest struct {
	Id                int
	Image1            *multipart.FileHeader `form:"image1"`
	Image2            *multipart.FileHeader `form:"image2"`
	Image3            *multipart.FileHeader `form:"image3"`
	Image4            *multipart.FileHeader `form:"image4"`
	Images            []string
	Name              string   `form:"name"`
	Description       string   `form:"description"`
	Price             *float64 `form:"price"`
	DiscountPercent   *float64 `form:"discount_percent"`
	Rating            *float64 `form:"rating"`
	IsFlashSale       bool     `form:"is_flash_sale"`
	Stock             *int     `form:"stock"`
	IsActive          bool     `form:"is_active"`
	SizeProducts      string   `form:"size_products"`
	ProductCategories string   `form:"product_categories"`
}

func TotalDataProducts(search string) (int, error) {
	var totalData int
	var err error
	searchParam := "%" + search + "%"
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(DISTINCT p.id)
			 FROM products p
			 LEFT JOIN product_categories pc ON pc.product_id = p.id
			 LEFT JOIN categories c ON c.id = pc.category_id
			 WHERE p.name ILIKE $1
			 OR p.description ILIKE $1
			 OR c.name ILIKE $1`, searchParam).Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM products`).Scan(&totalData)
	}

	return totalData, err
}

func GetListAllProducts(search string, page int, limit int) ([]Product, error) {
	var rows pgx.Rows
	var err error
	products := []Product{}
	offset := (page - 1) * limit
	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				p.is_active,
				COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			WHERE p.name ILIKE $3
				OR p.description ILIKE $3
				OR c.name ILIKE $3
			GROUP BY p.id
			ORDER BY p.id ASC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				p.is_active,
				COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			GROUP BY p.id
			ORDER BY p.id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return products, err
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return products, err
	}

	return products, err
}
