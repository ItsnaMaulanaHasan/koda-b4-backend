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
