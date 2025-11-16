package models

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
)

type AdminProductResponse struct {
	Id                int      `db:"id" json:"id"`
	ProductImages     []string `db:"product_images" json:"productImages"`
	Name              string   `db:"name" json:"name"`
	Description       string   `db:"description" json:"description"`
	Price             float64  `db:"price" json:"price"`
	DiscountPercent   float64  `db:"discount_percent" json:"discountPercent"`
	Rating            float64  `db:"rating" json:"rating"`
	IsFlashSale       bool     `db:"is_flash_sale" json:"isFlashSale"`
	Stock             int      `db:"stock" json:"stock"`
	IsActive          bool     `db:"is_active" json:"isActive"`
	IsFavourite       bool     `db:"is_favourite" json:"isFavourite"`
	ProductSizes      []string `db:"product_sizes" json:"productSizes"`
	ProductCategories []string `db:"product_categories" json:"productCategories"`
	ProductVariants   []string `db:"product_variants" json:"productVariants"`
}

type ProductRequest struct {
	Id                int
	FileImages        []*multipart.FileHeader `form:"fileImages"`
	ProductImages     []string
	Name              *string  `form:"name"`
	Description       *string  `form:"description"`
	Price             *float64 `form:"price"`
	DiscountPercent   *float64 `form:"discountPercent"`
	Rating            *float64 `form:"rating"`
	IsFlashSale       *bool    `form:"isFlashSale"`
	Stock             *int     `form:"stock"`
	IsActive          *bool    `form:"isActive"`
	IsFavourite       *bool    `form:"isFavourite"`
	SizeProducts      string   `form:"sizeProducts"`
	ProductCategories string   `form:"productCategories"`
	ProductVariants   string   `form:"productVariants"`
}

type PublicProductResponse struct {
	Id              int      `db:"id" json:"id"`
	ProductImages   []string `db:"product_images" json:"productImages"`
	Name            string   `db:"name" json:"name"`
	Description     string   `db:"description" json:"description"`
	Price           float64  `db:"price" json:"price"`
	DiscountPercent float64  `db:"discount_percent" json:"discountPercent"`
	IsFlashSale     bool     `db:"is_flash_sale" json:"isFlashSale"`
	IsFavourite     bool     `db:"is_favourite" json:"isFavourite"`
}

type PublicProductDetailResponse struct {
	Id                int                     `db:"id" json:"id"`
	ProductImages     []string                `db:"product_images" json:"productImages"`
	Name              string                  `db:"name" json:"name"`
	Description       string                  `db:"description" json:"description"`
	Price             float64                 `db:"price" json:"price"`
	DiscountPercent   float64                 `db:"discount_percent" json:"discountPercent"`
	Rating            float64                 `db:"rating" json:"rating"`
	IsFlashSale       bool                    `db:"is_flash_sale" json:"isFlashSale"`
	Stock             int                     `db:"stock" json:"stock"`
	ProductSizes      []string                `db:"product_sizes" json:"productSizes"`
	ProductCategories []string                `db:"product_categories" json:"productCategories"`
	ProductVariants   []string                `db:"product_variants" json:"productVariants"`
	Recomendations    []PublicProductResponse `db:"-" json:"Recomendations"`
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
				COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS product_sizes,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN product_sizes sp ON sp.product_id = p.id
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
				COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS product_sizes,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN product_sizes sp ON sp.product_id = p.id
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

func GetDetailProductAdmin(id int) (AdminProductResponse, string, error) {
	product := AdminProductResponse{}
	message := ""
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
				COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS product_sizes,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN product_sizes sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		message = "Failed to fetch detail product from database"
		return product, message, err
	}
	defer rows.Close()

	product, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[AdminProductResponse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Product not found"
			return product, message, err
		}
		message = "Failed to process detail product"
		return product, message, err
	}

	return product, message, nil
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

func InsertDataProduct(tx pgx.Tx, bodyCreate *ProductRequest, userIdFromToken any) error {
	err := tx.QueryRow(
		context.Background(),
		`INSERT INTO products (name, description, price, discount_percent, rating, is_flash_sale, stock, is_active, is_favourite, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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

func InsertProductImages(tx pgx.Tx, productId int, imagePaths []string, userId int) error {
	for i, imagePath := range imagePaths {
		isPrimary := (i == 0)
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO product_images (product_image, product_id, is_primary, created_by, updated_by)
			 VALUES ($1, $2, $3, $4, $5)`,
			imagePath,
			productId,
			isPrimary,
			userId,
			userId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertProductSizes(tx pgx.Tx, productId int, sizeIds []int, userId int) error {
	for _, sizeId := range sizeIds {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO product_sizes (product_id, size_id, created_by, updated_by)
			 VALUES ($1, $2, $3, $4)`,
			productId,
			sizeId,
			userId,
			userId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertProductCategories(tx pgx.Tx, productId int, categoryIds []int, userId int) error {
	for _, categoryId := range categoryIds {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO product_categories (product_id, category_id, created_by, updated_by)
			 VALUES ($1, $2, $3, $4)`,
			productId,
			categoryId,
			userId,
			userId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertProductVariants(tx pgx.Tx, productId int, variantIds []int, userId int) error {
	for _, variantId := range variantIds {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO product_variants (product_id, variant_id, created_by, updated_by)
			 VALUES ($1, $2, $3, $4)`,
			productId,
			variantId,
			userId,
			userId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteProductImages(tx pgx.Tx, productId int) error {
	_, err := tx.Exec(
		context.Background(),
		`DELETE FROM product_images WHERE product_id = $1`,
		productId,
	)
	return err
}

func DeleteProductSizes(tx pgx.Tx, productId int) error {
	_, err := tx.Exec(
		context.Background(),
		`DELETE FROM product_sizes WHERE product_id = $1`,
		productId,
	)
	return err
}

func DeleteProductCategories(tx pgx.Tx, productId int) error {
	_, err := tx.Exec(
		context.Background(),
		`DELETE FROM product_categories WHERE product_id = $1`,
		productId,
	)
	return err
}

func DeleteProductVariants(tx pgx.Tx, productId int) error {
	_, err := tx.Exec(
		context.Background(),
		`DELETE FROM product_variants WHERE product_id = $1`,
		productId,
	)
	return err
}

func UpdateDataProduct(tx pgx.Tx, productId int, bodyUpdate *ProductRequest, userId int) error {
	_, err := tx.Exec(
		context.Background(),
		`UPDATE products 
		 SET name             = COALESCE($1, name),
		     description      = COALESCE($2, description),
		     price            = COALESCE($3, price),
		     discount_percent = COALESCE($4, discount_percent),
		     stock            = COALESCE($5, stock),
		     is_flash_sale    = COALESCE($6, is_flash_sale),
		     is_active        = COALESCE($7, is_active),
		     is_favourite     = COALESCE($8, is_favourite),
		     updated_by       = $9,
		     updated_at       = NOW()
		 WHERE id = $10`,
		bodyUpdate.Name,
		bodyUpdate.Description,
		bodyUpdate.Price,
		bodyUpdate.DiscountPercent,
		bodyUpdate.Stock,
		bodyUpdate.IsFlashSale,
		bodyUpdate.IsActive,
		bodyUpdate.IsFavourite,
		userId,
		productId,
	)
	return err
}

func DeleteDataProduct(productId int) (bool, string, error) {
	isSuccess := false
	message := ""

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	// check if product exists
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", productId).Scan(&exists)
	if err != nil {
		message = "Internal server error while checking product existence"
		return isSuccess, message, err
	}

	if !exists {
		message = "Product not found"
		return isSuccess, message, nil
	}

	// delete product (CASCADE will auto delete related data)
	commandTag, err := tx.Exec(ctx, `DELETE FROM products WHERE id = $1`, productId)
	if err != nil {
		message = "Internal server error while deleting product data"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Product not found"
		return isSuccess, message, nil
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "Product deleted successfully"
	return isSuccess, message, nil
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
			COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images
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
			COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images
		FROM products p
		LEFT JOIN product_images pi ON pi.product_id = p.id
		LEFT JOIN product_categories pc ON pc.product_id = p.id
		LEFT JOIN categories c ON c.id = pc.category_id
		WHERE p.is_active = true`

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
	orderBy := "p.id ASC"
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

func GetDetailProductPublic(id int) (PublicProductDetailResponse, string, error) {
	product := PublicProductDetailResponse{}
	message := ""
	query := `SELECT 
				p.id,
				p.name,
				p.description,
				p.price,
				COALESCE(p.discount_percent, 0) AS discount_percent,
				COALESCE(p.rating, 0) AS rating,
				p.is_flash_sale,
				COALESCE(p.stock, 0) AS stock,
				COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images,
				COALESCE(ARRAY_AGG(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS product_sizes,
				COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') AS product_categories,
				COALESCE(ARRAY_AGG(DISTINCT v.name) FILTER (WHERE v.name IS NOT NULL), '{}') AS product_variants
			FROM products p
			LEFT JOIN product_images pi ON pi.product_id = p.id
			LEFT JOIN product_sizes sp ON sp.product_id = p.id
			LEFT JOIN sizes s ON s.id = sp.size_id
			LEFT JOIN product_categories pc ON pc.product_id = p.id
			LEFT JOIN categories c ON c.id = pc.category_id
			LEFT JOIN product_variants pv ON pv.product_id = p.id
			LEFT JOIN variants v ON v.id = pv.variant_id
			WHERE p.id = $1
			GROUP BY p.id;`

	rows, err := config.DB.Query(context.Background(), query, id)
	if err != nil {
		message = "Failed to fetch detail product from database"
		return product, message, err
	}
	defer rows.Close()

	product, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[PublicProductDetailResponse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Product not found"
			return product, message, err
		}
		message = "Failed tp process detail product"
		return product, message, err
	}

	queryRecomendation := `
						SELECT 
							p.id,
							p.name,
							p.description,
							p.price,
							COALESCE(p.discount_percent, 0) AS discount_percent,
							p.is_flash_sale,
							p.is_favourite,
							COALESCE(ARRAY_AGG(DISTINCT pi.product_image) FILTER (WHERE pi.product_image IS NOT NULL), '{}') AS product_images
						FROM products p
						LEFT JOIN product_images pi ON pi.product_id = p.id
						WHERE p.is_active = true
						AND p.id != $1
						AND EXISTS (
							SELECT 1
							FROM product_categories pc1
							WHERE pc1.product_id = p.id
							AND pc1.category_id IN (
								SELECT pc2.category_id
								FROM product_categories pc2
								WHERE pc2.product_id = $1
							)
						)
						GROUP BY p.id
						ORDER BY RANDOM()
						LIMIT 5;`

	rowsRec, err := config.DB.Query(context.Background(), queryRecomendation, id)
	if err != nil {
		message = "Failed to get recomendation product from database"
		return product, message, err
	}
	defer rowsRec.Close()

	product.Recomendations, _ = pgx.CollectRows(rowsRec, pgx.RowToStructByName[PublicProductResponse])

	return product, message, nil

}

func InvalidateProductCache(ctx context.Context) error {
	rdb := lib.Redis()

	patterns := []string{
		"products:total:*",
		"products:list:*",
		"products:detail:*",
	}

	for _, pattern := range patterns {
		keys, err := rdb.Keys(ctx, pattern).Result()
		if err != nil {
			log.Printf("Failed to get keys for pattern %s: %v", pattern, err)
			continue
		}

		if len(keys) > 0 {
			err := rdb.Del(ctx, keys...).Err()
			if err != nil {
				log.Printf("Failed to delete keys for pattern %s: %v", pattern, err)
			} else {
				log.Printf("Invalidated %d cache keys for pattern %s", len(keys), pattern)
			}
		}
	}

	return nil
}
