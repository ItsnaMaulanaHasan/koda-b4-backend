package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
)

type Product struct {
	Id                int      `db:"id" json:"id"`
	Images            []string `db:"images" json:"images"`
	Name              string   `db:"name" json:"name"`
	Description       string   `db:"description" json:"description"`
	Price             float64  `db:"price" json:"price"`
	DiscountPercent   float64  `db:"discount_percent" json:"discountPercent"`
	Rating            float64  `db:"rating" json:"rating"`
	IsFlashSale       bool     `db:"is_flash_sale" json:"isFlashSale"`
	Stock             int      `db:"stock" json:"stock"`
	IsActive          bool     `db:"is_active" json:"isActive"`
	IsFavourite       bool     `db:"is_favourite" json:"isFavourite"`
	SizeProducts      []string `db:"size_products" json:"sizeProducts"`
	ProductCategories []string `db:"product_categories" json:"productCategories"`
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
	DiscountPercent   *float64 `form:"discountPercent"`
	Rating            *float64 `form:"rating"`
	IsFlashSale       bool     `form:"isFlashSale"`
	Stock             *int     `form:"stock"`
	IsActive          bool     `form:"isActive"`
	IsFavourite       bool     `form:"isFavourite"`
	SizeProducts      string   `form:"sizeProducts"`
	ProductCategories string   `form:"productCategories"`
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
				p.is_favourite,
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
				p.is_favourite,
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

	return products, nil
}

func GetProductById(id int) (Product, error) {
	product := Product{}
	query := `SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				p.is_active,
				p.is_favourite,
				COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id 
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		return product, err
	}
	defer rows.Close()

	product, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product, err
		}
		return product, err
	}

	return product, nil
}

func CheckProductName(name string) (bool, error) {
	exists := false
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)", name,
	).Scan(&exists)
	if err != nil {
		return exists, err
	}

	return exists, nil
}

func InsertDataProduct(bodyCreate *ProductRequest, userIdFromToken any) error {
	err := config.DB.QueryRow(
		context.Background(),
		`INSERT INTO products (name, description, price, discount_percent, rating, is_flash_sale, stock, is_active, is_favourite, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		bodyCreate.Name,
		bodyCreate.Description,
		bodyCreate.Price,
		bodyCreate.DiscountPercent,
		bodyCreate.Rating,
		bodyCreate.IsFlashSale,
		bodyCreate.Stock,
		bodyCreate.IsActive,
		bodyCreate.IsFavourite,
		userIdFromToken,
		userIdFromToken,
	).Scan(&bodyCreate.Id)
	if err != nil {
		return err
	}

	return nil
}

func GetListFavouriteProducts(limit int) ([]Product, error) {
	var rows pgx.Rows
	var err error
	products := []Product{}
	rows, err = config.DB.Query(context.Background(),
		`SELECT 
			p.id,
			p.name,
			p.description,
			p.price,
			COALESCE(p.discount_percent, 0) AS discount_percent,
			COALESCE(p.rating, 0) AS rating,
			p.is_flash_sale,
			p.is_active,
			p.is_favourite
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
			COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
			COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories
		FROM products p
		LEFT JOIN product_images pi ON pi.product_id = p.id
		LEFT JOIN size_products sp ON sp.product_id = p.id
		LEFT JOIN sizes s ON s.id = sp.size_id
		LEFT JOIN product_categories pc ON pc.product_id = p.id
		LEFT JOIN categories c ON c.id = pc.category_id
		WHERE p.is_favourite = true AND p.is_active = true
		GROUP BY p.id
		ORDER BY p.id ASC
		LIMIT $1 OFFSET`, limit)

	if err != nil {
		return products, err
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return products, err
	}

	return products, nil
}
