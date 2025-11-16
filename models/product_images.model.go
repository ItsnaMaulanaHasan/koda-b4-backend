package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductImage struct {
	Id           int       `json:"id" db:"id"`
	ProductId    int       `json:"productId" db:"product_id"`
	ProductImage string    `json:"productImage" db:"product_image"`
	IsPrimary    bool      `json:"isPrimary" db:"is_primary"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type ProductImageRequest struct {
	Image     *multipart.FileHeader `form:"image"`
	IsPrimary bool                  `form:"isPrimary"`
}

func CheckProductExists(productId int) (bool, error) {
	var exists bool
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)",
		productId,
	).Scan(&exists)
	return exists, err
}

func GetProductImages(productId int) ([]ProductImage, string, error) {
	images := []ProductImage{}
	message := ""

	rows, err := config.DB.Query(
		context.Background(),
		`SELECT id, product_id, product_image, is_primary, created_at, updated_at
		 FROM product_images
		 WHERE product_id = $1
		 ORDER BY is_primary DESC, id ASC`,
		productId,
	)
	if err != nil {
		message = "Failed to fetch product images from database"
		return images, message, err
	}
	defer rows.Close()

	images, err = pgx.CollectRows(rows, pgx.RowToStructByName[ProductImage])
	if err != nil {
		message = "Failed to process product images data"
		return images, message, err
	}

	message = "Success get product images"
	return images, message, nil
}

func GetProductImageById(imageId int) (ProductImage, string, error) {
	image := ProductImage{}
	message := ""

	rows, err := config.DB.Query(
		context.Background(),
		`SELECT id, product_id, product_image, is_primary, created_at, updated_at
		 FROM product_images
		 WHERE id = $1`,
		imageId,
	)
	if err != nil {
		message = "Failed to fetch product image from database"
		return image, message, err
	}
	defer rows.Close()

	image, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[ProductImage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Product image not found"
			return image, message, err
		}
		message = "Failed to process product image data"
		return image, message, err
	}

	message = "Success get product image"
	return image, message, nil
}

func InsertProductImage(productId int, imagePath string, isPrimary bool, userId int) (int, string, error) {
	var imageId int
	message := ""

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return imageId, message, err
	}
	defer tx.Rollback(ctx)

	// If this image is primary, set all other images to non-primary
	if isPrimary {
		_, err = tx.Exec(
			ctx,
			`UPDATE product_images 
			 SET is_primary = false 
			 WHERE product_id = $1`,
			productId,
		)
		if err != nil {
			message = "Internal server error while updating other images"
			return imageId, message, err
		}
	}

	err = tx.QueryRow(
		ctx,
		`INSERT INTO product_images (product_image, product_id, is_primary, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		imagePath,
		productId,
		isPrimary,
		userId,
		userId,
	).Scan(&imageId)
	if err != nil {
		message = "Internal server error while inserting product image"
		return imageId, message, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return imageId, message, err
	}

	message = "Product image created successfully"
	return imageId, message, nil
}

func UpdateProductImage(imageId int, isPrimary bool, userId int) (bool, string, error) {
	isSuccess := false
	message := ""

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return isSuccess, message, err
	}
	defer tx.Rollback(ctx)

	// Get product_id from image
	var productId int
	err = tx.QueryRow(ctx, `SELECT product_id FROM product_images WHERE id = $1`, imageId).Scan(&productId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "Product image not found"
			return isSuccess, message, nil
		}
		message = "Internal server error while fetching product image"
		return isSuccess, message, err
	}

	// If setting as primary, unset all other images
	if isPrimary {
		_, err = tx.Exec(
			ctx,
			`UPDATE product_images 
			 SET is_primary = false 
			 WHERE product_id = $1 AND id != $2`,
			productId,
			imageId,
		)
		if err != nil {
			message = "Internal server error while updating other images"
			return isSuccess, message, err
		}
	}

	commandTag, err := tx.Exec(
		ctx,
		`UPDATE product_images 
		 SET is_primary = $1,
		     updated_by = $2,
		     updated_at = NOW()
		 WHERE id = $3`,
		isPrimary,
		userId,
		imageId,
	)
	if err != nil {
		message = "Internal server error while updating product image"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Product image not found"
		return isSuccess, message, nil
	}

	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return isSuccess, message, err
	}

	isSuccess = true
	message = "Product image updated successfully"
	return isSuccess, message, nil
}

func DeleteProductImage(imageId int) (bool, string, error) {
	isSuccess := false
	message := ""

	commandTag, err := config.DB.Exec(
		context.Background(),
		`DELETE FROM product_images WHERE id = $1`,
		imageId,
	)
	if err != nil {
		message = "Internal server error while deleting product image"
		return isSuccess, message, err
	}

	if commandTag.RowsAffected() == 0 {
		message = "Product image not found"
		return isSuccess, message, nil
	}

	isSuccess = true
	message = "Product image deleted successfully"
	return isSuccess, message, nil
}
