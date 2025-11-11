package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
)

type AdminProductResponse struct {
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
	ProductVariants   []string `db:"product_variants" json:"productVariants"`
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
	ProductVariants   string   `form:"productVariants"`
}

type PublicProductResponse struct {
	Id                int      `db:"id" json:"id"`
	Images            []string `db:"images" json:"images"`
	Name              string   `db:"name" json:"name"`
	Description       string   `db:"description" json:"description"`
	Price             float64  `db:"price" json:"price"`
	DiscountPercent   float64  `db:"discount_percent" json:"discountPercent"`
	IsFlashSale       bool     `db:"is_flash_sale" json:"isFlashSale"`
	IsFavourite       bool     `db:"is_favourite" json:"isFavourite"`
	ProductCategories []string `db:"product_categories" json:"productCategories"`
}

type PublicProductDetailResponse struct {
	Id                int      `db:"id" json:"id"`
	Images            []string `db:"images" json:"images"`
	Name              string   `db:"name" json:"name"`
	Description       string   `db:"description" json:"description"`
	Price             float64  `db:"price" json:"price"`
	DiscountPercent   float64  `db:"discount_percent" json:"discountPercent"`
	Rating            float64  `db:"rating" json:"rating"`
	IsFlashSale       bool     `db:"is_flash_sale" json:"isFlashSale"`
	Stock             int      `db:"stock" json:"stock"`
	SizeProducts      []string `db:"size_products" json:"sizeProducts"`
	ProductCategories []string `db:"product_categories" json:"productCategories"`
	ProductVariants   []string `db:"product_variants" json:"productVariants"`
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

func GetListProductsAdmin(search string, page int, limit int) ([]AdminProductResponse, error) {
	var rows pgx.Rows
	var err error
	products := []AdminProductResponse{}
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
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
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
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
			GROUP BY p.id
			ORDER BY p.id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return products, err
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[AdminProductResponse])
	if err != nil {
		return products, err
	}

	return products, nil
}

func GetProductByIdAdmin(id int) (AdminProductResponse, error) {
	product := AdminProductResponse{}
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
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		return product, err
	}
	defer rows.Close()

	product, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[AdminProductResponse])
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

func GetListFavouriteProducts(limit int) ([]PublicProductResponse, error) {
	var rows pgx.Rows
	var err error
	products := []PublicProductResponse{}
	rows, err = config.DB.Query(context.Background(),
		`SELECT 
			p.id,
			p.name,
			p.description,
			p.price,
			COALESCE(p.discount_percent, 0) AS discount_percent,
			p.is_flash_sale,
			p.is_favourite,
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
		FROM products p
		LEFT JOIN product_images pi ON pi.product_id = p.id
		WHERE p.is_favourite = true AND p.is_active = true
		GROUP BY p.id
		ORDER BY p.id ASC
		LIMIT $1`, limit)

	if err != nil {
		return products, err
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[PublicProductResponse])
	if err != nil {
		return products, err
	}

	return products, nil
}

func GetListProductsPublic(q string, cat []string, sort string, maxPrice float64, minPrice float64, limit int, page int) ([]PublicProductResponse, error) {
	products := []PublicProductResponse{}
	offset := (page - 1) * limit

	query := `
		SELECT 
			p.id,
			p.name,
			p.description,
			p.price,
			COALESCE(p.discount_percent, 0) AS discount_percent,
			p.is_flash_sale,
			p.is_favourite,
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
		FROM products p
		LEFT JOIN product_images pi ON pi.product_id = p.id
		WHERE p.is_active = true
		GROUP BY p.id
		ORDER BY p.id ASC
		WHERE 1=1`

	// Dynamic parameters
	args := []any{}
	paramCount := 1

	// Search filter
	if q != "" {
		query += fmt.Sprintf(` AND p.name ILIKE $%d`, paramCount)
		args = append(args, "%"+q+"%")
		paramCount++
	}

	// Category filter
	if len(cat) > 0 {
		query += fmt.Sprintf(` AND c.name = ANY($%d)`, paramCount)
		args = append(args, cat)
		paramCount++
	}

	// Price range filter
	if minPrice > 0 {
		query += fmt.Sprintf(` AND p.price >= $%d`, paramCount)
		args = append(args, minPrice)
		paramCount++
	}

	if maxPrice > 0 {
		query += fmt.Sprintf(` AND p.price <= $%d`, paramCount)
		args = append(args, maxPrice)
		paramCount++
	}

	// Group by
	query += ` GROUP BY p.id`

	// Sorting
	orderBy := "p.id ASC" // default
	if sort != "" {
		switch sort {
		case "name_asc":
			orderBy = "p.name ASC"
		case "name_desc":
			orderBy = "p.name DESC"
		case "price_asc":
			orderBy = "p.price ASC"
		case "price_desc":
			orderBy = "p.price DESC"
		}
	}
	query += fmt.Sprintf(` ORDER BY %s`, orderBy)

	// Pagination
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, paramCount, paramCount+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := config.DB.Query(context.Background(), query, args...)
	if err != nil {
		return products, err
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[PublicProductResponse])
	if err != nil {
		return products, err
	}

	return products, nil
}

func GetProductByIdPublic(id int) (PublicProductDetailResponse, error) {
	product := PublicProductDetailResponse{}
	query := `SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS size_products,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN size_products sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		return product, err
	}
	defer rows.Close()

	product, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[PublicProductDetailResponse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product, err
		}
		return product, err
	}

	return product, nil
}
