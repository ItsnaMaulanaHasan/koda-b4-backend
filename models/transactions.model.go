package models

import (
	"backend-daily-greens/config"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Transaction struct {
	Id              int       `db:"id"`
	NoInvoice       string    `db:"no_invoice"`
	DateTransaction time.Time `db:"date_transaction"`
	Status          string    `db:"status"`
	Total           float64   `db:"total_transaction"`
}

type TransactionDetail struct {
	Id               int                `json:"id" db:"id"`
	UserId           int                `json:"userId" db:"user_id"`
	NoInvoice        string             `json:"noInvoice" db:"no_invoice"`
	DateTransaction  time.Time          `json:"dateOrder" db:"date_transaction"`
	FullName         string             `json:"fullName" db:"full_name"`
	Email            string             `json:"email" db:"email"`
	Address          string             `json:"address" db:"address"`
	Phone            string             `json:"phone" db:"phone"`
	PaymentMethod    string             `json:"payment_method" db:"payment_method"`
	OrderMethod      string             `json:"orderMethod" db:"order_method"`
	Status           string             `json:"status" db:"status"`
	DeliveryFee      float64            `json:"delivery_fee" db:"delivery_fee"`
	AdminFee         float64            `json:"adminFee" db:"admin_fee"`
	Tax              float64            `json:"tax" db:"tax"`
	TotalTransaction float64            `json:"totalTransaction" db:"total_transaction"`
	TransactionItems []TransactionItems `json:"transactionItems"`
}

type TransactionItems struct {
	Id              int     `json:"id" db:"id"`
	TransactionId   int     `json:"transactionId" db:"transaction_id"`
	ProductId       int     `json:"product_id" db:"product_id"`
	ProductName     string  `json:"product_name" db:"product_name"`
	ProductPrice    float64 `json:"product_price" db:"product_price"`
	DiscountPercent float64 `json:"discount_percent" db:"discount_percent"`
	DiscountPrice   float64 `json:"discount_price" db:"discount_price"`
	Size            string  `json:"size" db:"size"`
	SizeCost        float64 `json:"sizeCost" db:"size_cost"`
	Variant         string  `json:"variant" db:"variant"`
	VariantCost     float64 `json:"variantCost" db:"variant_cost"`
	Amount          int     `json:"amount" db:"amount"`
	Subtotal        float64 `json:"subtotal" db:"subtotal"`
}

type TransactionRequest struct {
	NoInvoice        string    `json:"-" swaggerignore:"true"`
	DateTransaction  time.Time `json:"-" swaggerignore:"true"`
	FullName         string    `json:"fullName"`
	Email            string    `json:"email"`
	Address          string    `json:"address"`
	Phone            string    `json:"phone"`
	PaymentMethodId  int       `json:"paymentMethodId" binding:"required"`
	OrderMethodId    int       `json:"orderMethodId" binding:"required"`
	DeliveryFee      float64   `json:"-" swaggerignore:"true"`
	AdminFee         float64   `json:"-" swaggerignore:"true"`
	Tax              float64   `json:"-" swaggerignore:"true"`
	TotalTransaction float64   `json:"-" swaggerignore:"true"`
}

func GetTotalDataTransactions(search string) (int, error) {
	totalData := 0
	var err error
	if search != "" {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) 
			 FROM transactions
			 WHERE no_invoice ILIKE $1`, "%"+search+"%").Scan(&totalData)
	} else {
		err = config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM transactions`).Scan(&totalData)
	}
	if err != nil {
		return totalData, err
	}

	return totalData, nil
}

func GetListAllTransactions(page int, limit int, search string) ([]Transaction, string, error) {
	offset := (page - 1) * limit
	var rows pgx.Rows
	var err error
	message := ""
	transactions := []Transaction{}

	if search != "" {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				id,
				no_invoice,
				date_transaction,
				status,
				total_transaction
			FROM transactions
			WHERE no_invoice ILIKE $3
			ORDER BY date_transaction DESC, id DESC
			LIMIT $1 OFFSET $2`, limit, offset, "%"+search+"%")
	} else {
		rows, err = config.DB.Query(context.Background(),
			`SELECT 
				id,
				no_invoice,
				date_transaction,
				status,
				total_transaction
			FROM transactions
			ORDER BY date_transaction DESC, id DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		message = "Failed to fetch transactions from database"
		return transactions, message, err
	}
	defer rows.Close()

	transactions, err = pgx.CollectRows(rows, pgx.RowToStructByName[Transaction])
	if err != nil {
		message = "Failed to process transaction data from database"
		return transactions, message, err
	}

	message = "Success get all transaction"
	return transactions, message, nil
}

func GetDeliveryFeeAndAdminFee(orderMethodId int, paymentMethodId int) (float64, float64, string, error) {
	var deliveryFee, adminFee float64 = 0, 0
	message := ""

	// get delivery fee by order method id
	err := config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(delivery_fee, 0) FROM order_methods WHERE id = $1`,
		orderMethodId,
	).Scan(&deliveryFee)
	if err != nil {
		message = "Invalid order method id"
		return deliveryFee, adminFee, message, err
	}

	// get admin fee by payment method id
	err = config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(admin_fee, 0) FROM payment_methods WHERE id = $1`,
		paymentMethodId,
	).Scan(&adminFee)
	if err != nil {
		message = "Invalid payment method id"
		return deliveryFee, adminFee, message, err
	}

	return deliveryFee, adminFee, message, nil
}

func MakeTransaction(userId int, bodyCheckout TransactionRequest, carts []Cart) (int, string, error) {
	message := ""
	ctx := context.Background()
	// start transaction
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start transaction"
		return 0, message, err
	}
	defer tx.Rollback(ctx)

	// insert data to transactions
	var transactionId int
	insertTransaction := `INSERT INTO transactions (
							user_id, 
							no_invoice, 
							date_transaction,
							full_name,
							email,
							address,
							phone,
							payment_method_id,
							order_method_id,
							delivery_fee,
							admin_fee,
							tax,
							total_transaction,
							created_by,
							updated_by)
						VALUES 
							($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
						RETURNING 
							id`

	err = tx.QueryRow(ctx, insertTransaction,
		userId,
		bodyCheckout.NoInvoice,
		bodyCheckout.DateTransaction,
		bodyCheckout.FullName,
		bodyCheckout.Email,
		bodyCheckout.Address,
		bodyCheckout.Phone,
		bodyCheckout.PaymentMethodId,
		bodyCheckout.OrderMethodId,
		bodyCheckout.DeliveryFee,
		bodyCheckout.AdminFee,
		bodyCheckout.Tax,
		bodyCheckout.TotalTransaction,
		userId,
		userId,
	).Scan(&transactionId)
	if err != nil {
		message = "Failed to insert transaction"
		return 0, message, err
	}

	// insert data to transaction_items
	for _, cart := range carts {
		queryOrdered := `INSERT INTO transaction_items (
							transaction_id, 
							product_id, 
							product_name, 
							product_price, 
							discount_percent, 
							discount_price, 
							size, 
							size_cost, 
							variant, 
							variant_cost, 
							amount, 
							subtotal, 
							created_by, 
							updated_by) 
						VALUES 
							($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`

		_, err := tx.Exec(ctx, queryOrdered,
			transactionId,
			cart.ProductId,
			cart.ProductName,
			cart.ProductPrice,
			cart.DiscountPercent,
			cart.DiscountPrice,
			cart.SizeName,
			cart.SizeCost,
			cart.VariantName,
			cart.VariantCost,
			cart.Amount,
			cart.Subtotal,
			userId,
			userId,
		)
		if err != nil {
			message = "Failed to insert ordered product"
			return 0, message, err
		}

		// update stock
		_, err = tx.Exec(ctx,
			`UPDATE products SET stock = stock - $1 WHERE id = $2`,
			cart.Amount, cart.ProductId,
		)
		if err != nil {
			message = "Failed to update stock of product"
			return 0, message, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return 0, message, err
	}

	message = "Checkout completed successfully"
	return transactionId, message, nil
}
