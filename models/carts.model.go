package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Cart struct {
	Id              int      `db:"id" json:"id"`
	UserId          int      `db:"user_id" json:"userId"`
	ProductId       int      `db:"product_id" json:"productId"`
	ProductImages   []string `db:"product_images" json:"productImages"`
	ProductName     string   `db:"product_name" json:"productName"`
	ProductPrice    float64  `db:"product_price" json:"productPrice"`
	IsFlashSale     bool     `db:"is_flash_sale" json:"isFlashSale"`
	DiscountPercent float64  `db:"discount_percent" json:"discountPercent"`
	DiscountPrice   float64  `db:"discount_price" json:"discountPrice"`
	SizeName        string   `db:"size_name" json:"sizeName"`
	SizeCost        float64  `db:"size_cost" json:"sizeCost"`
	VariantName     string   `db:"variant_name" json:"variantName"`
	VariantCost     float64  `db:"variant_cost" json:"variantCost"`
	Amount          int      `db:"amount" json:"amount"`
	Subtotal        float64  `db:"subtotal" json:"subtotal"`
}

type CartRequest struct {
	Id        int     `json:"id" swaggerignore:"true"`
	UserId    int     `json:"userId" swaggerignore:"true"`
	ProductId int     `json:"productId"`
	SizeId    int     `json:"sizeId"`
	VariantId int     `json:"variantId"`
	Amount    int     `json:"amount"`
	Subtotal  float64 `json:"subtotal" swaggerignore:"true"`
}

func GetListCart(userId int) ([]Cart, string, error) {
	carts := []Cart{}
	message := ""
	var err error

	// get cart list
	rows, err := config.DB.Query(context.Background(),
		`SELECT 
			c.id, 
			c.user_id,
			c.product_id, 
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS product_images, 
			p.name AS product_name,
			p.price AS product_price,
			p.is_flash_sale,
			p.discount_percent,
			(p.price * (1-(p.discount_percent/100))) AS discount_price,
			s.name AS size_name,
			s.size_cost AS size_cost,
			v.name AS variant_name,
			v.variant_cost AS variant_cost,
			c.amount, 
			c.subtotal
			FROM carts c
		LEFT JOIN products p ON p.id = c.product_id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		LEFT JOIN sizes s  ON s.id = c.size_id
		LEFT JOIN variants v ON v.id = c.variant_id
		WHERE c.user_id = $1
		GROUP BY c.id, c.product_id, p.name, p.price, p.is_flash_sale, p.discount_percent, s.name, s.size_cost, v.name, v.variant_cost
		ORDER BY c.updated_at DESC`, userId)
	if err != nil {
		message = "Failed to fetch list carts from database"
		return carts, message, err
	}
	defer rows.Close()

	carts, err = pgx.CollectRows(rows, pgx.RowToStructByName[Cart])
	if err != nil {
		message = "Failed to process carts data from database"
		return carts, message, err
	}

	message = "Success get list carts"
	return carts, message, nil
}

func AddToCart(bodyAdd CartRequest) (CartRequest, string, error) {
	ctx := context.Background()
	responseCart := CartRequest{}
	message := ""

	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return responseCart, message, err
	}
	defer tx.Rollback(ctx)

	// get stock product
	var stock int
	err = tx.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1`, bodyAdd.ProductId).Scan(&stock)
	if err != nil {
		message = "Internal server error while get stock from products"
		return responseCart, message, err
	}

	if bodyAdd.Amount <= 0 {
		message = "invalid amount, must be greater than 0"
		return responseCart, message, errors.New(message)
	}

	if bodyAdd.Amount > stock {
		message = "amount exceeds available stock"
		return responseCart, message, errors.New(message)
	}

	// check whether the cart item already exists in the database
	var cartIsExist bool
	err = tx.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM carts WHERE user_id = $1 AND product_id = $2 AND size_id = $3 AND variant_id = $4)", bodyAdd.UserId, bodyAdd.ProductId, bodyAdd.SizeId, bodyAdd.VariantId).Scan(&cartIsExist)
	if err != nil {
		message = "Internal server error while checking cart"
		return responseCart, message, err
	}

	if cartIsExist {
		// get the previous amount
		var oldAmount int
		err = tx.QueryRow(ctx,
			`SELECT amount FROM carts 
     		WHERE user_id = $1 AND product_id = $2 AND size_id = $3 AND variant_id = $4`,
			bodyAdd.UserId, bodyAdd.ProductId, bodyAdd.SizeId, bodyAdd.VariantId,
		).Scan(&oldAmount)
		if err != nil {
			message = "Internal server error while get amount of the product"
			return responseCart, message, err
		}
		bodyAdd.Amount += oldAmount

		// calculate new subtotal
		err := tx.QueryRow(ctx,
			`SELECT 
				((p.price * (1-(p.discount_percent/100))) + s.size_cost + v.variant_cost) * $4 AS subtotal
			FROM products p
			JOIN sizes s ON s.id = $2
			JOIN variants v ON v.id = $3
			WHERE p.id = $1`,
			bodyAdd.ProductId, bodyAdd.SizeId, bodyAdd.VariantId, bodyAdd.Amount,
		).Scan(&bodyAdd.Subtotal)
		if err != nil {
			message = "Internal server error while calculate subtotal"
			return responseCart, message, err
		}

		// update cart items
		_, err = tx.Exec(
			ctx,
			`UPDATE carts SET amount = $1, subtotal = $2, updated_at = NOW(), updated_by = $3 WHERE user_id = $4`,
			bodyAdd.Amount,
			bodyAdd.Subtotal,
			bodyAdd.UserId,
			bodyAdd.UserId,
		)
		if err != nil {
			message = "Internal server error while updating cart"
			return responseCart, message, err
		}
	} else {
		// calculate subtotal for new cart
		err := tx.QueryRow(ctx,
			`SELECT 
				((p.price * (1-(p.discount_percent/100))) + s.size_cost + v.variant_cost) * $4 AS subtotal
			FROM products p
			JOIN sizes s ON s.id = $2
			JOIN variants v ON v.id = $3
			WHERE p.id = $1`,
			bodyAdd.ProductId, bodyAdd.SizeId, bodyAdd.VariantId, bodyAdd.Amount,
		).Scan(&bodyAdd.Subtotal)
		if err != nil {
			message = "Internal server error while calculate subtotal"
			return responseCart, message, err
		}

		// add cart items
		err = tx.QueryRow(
			ctx,
			`INSERT INTO carts (user_id, product_id, size_id, variant_id, amount, subtotal, created_by, updated_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id`,
			bodyAdd.UserId,
			bodyAdd.ProductId,
			bodyAdd.SizeId,
			bodyAdd.VariantId,
			bodyAdd.Amount,
			bodyAdd.Subtotal,
			bodyAdd.UserId,
			bodyAdd.UserId,
		).Scan(&bodyAdd.Id)
		if err != nil {
			message = "Internal server error while adding cart"
			return responseCart, message, err
		}
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return responseCart, message, err
	}

	message = "Cart added successfully"
	responseCart = CartRequest{
		Id:        bodyAdd.Id,
		UserId:    bodyAdd.UserId,
		ProductId: bodyAdd.ProductId,
		SizeId:    bodyAdd.SizeId,
		VariantId: bodyAdd.VariantId,
		Amount:    bodyAdd.Amount,
		Subtotal:  bodyAdd.Subtotal,
	}

	return responseCart, message, nil
}

func DeleteCartById(cartId int) (pgconn.CommandTag, error) {
	commandTag, err := config.DB.Exec(context.Background(), `DELETE FROM carts WHERE id = $1`, cartId)
	if err != nil {
		return commandTag, err
	}

	return commandTag, nil
}
