package models

import (
	"backend-daily-greens/config"
	"context"

	"github.com/jackc/pgx/v5"
)

type Cart struct {
	Id            int      `db:"id" json:"id"`
	UserId        int      `db:"user_id" json:"userId"`
	ProductImages []string `db:"product_images" json:"productImages"`
	ProductName   string   `db:"product_name" json:"productName"`
	SizeName      string   `db:"size_name" json:"sizeName"`
	VariantName   string   `db:"variant_name" json:"variantName"`
	ProductPrice  string   `db:"product_price" json:"productPrice"`
	SizeCost      string   `db:"size_cost" json:"sizeCost"`
	VariantCost   string   `db:"variant_cost" json:"variantCost"`
	Amount        int      `db:"amount" json:"amount"`
	Subtotal      float64  `db:"subtotal" json:"subtotal"`
}

type CartRequest struct {
	Id        int     `json:"-"`
	UserId    int     `json:"-"`
	ProductId int     `json:"productId"`
	SizeId    int     `json:"sizeId"`
	VariantId int     `json:"variantId"`
	Amount    float64 `json:"amount"`
	Subtotal  float64 `json:"-"`
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
			COALESCE(ARRAY_AGG(DISTINCT pi.image) FILTER (WHERE pi.image IS NOT NULL), '{}') AS product_images, 
			p.name AS product_name,
			s.name AS size_name,
			v.name AS variant_name, 
			p.price AS product_price,
			v.variant_cost AS variant_cost,
			s.size_cost AS size_cost,
			c.amount, 
			c.subtotal
			FROM carts c
		LEFT JOIN products p ON p.id = c.product_id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		LEFT JOIN sizes s  ON s.id = c.size_id
		LEFT JOIN variants v ON v.id = c.variant_id
		WHERE c.user_id = $1
		GROUP BY c.id, p.name, s.name, v.name, p.price, s.size_cost, v.variant_cost
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
	var cartIsExist bool
	responseCart := CartRequest{}
	message := ""

	// check whether the cart item already exists in the database
	err := config.DB.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM carts WHERE user_id = $1 AND product_id = $2 AND size_id = $3 AND variant_id = $4)", bodyAdd.UserId, bodyAdd.ProductId, bodyAdd.SizeId, bodyAdd.VariantId).Scan(&cartIsExist)
	if err != nil {
		message = "Internal server error while checking cart"
		return responseCart, message, err
	}

	if cartIsExist {
		var oldAmount float64

		// get the previous amount
		err = config.DB.QueryRow(context.Background(), `SELECT amount FROM carts WHERE user_id = $1`, bodyAdd.UserId).Scan(&oldAmount)
		if err != nil {
			message = "Internal server error while get amount of the product"
			return responseCart, message, err
		}
		bodyAdd.Amount += oldAmount

		//get the new subtotal
		err := config.DB.QueryRow(context.Background(),
			`SELECT 
				(p.price + s.size_cost + v.variant_cost) * $4 AS subtotal
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
		_, err = config.DB.Exec(
			context.Background(),
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
		// get the subtotal
		err := config.DB.QueryRow(context.Background(),
			`SELECT 
				(p.price + s.size_cost + v.variant_cost) * $4 AS subtotal
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
		err = config.DB.QueryRow(
			context.Background(),
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
