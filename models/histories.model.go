package models

import (
	"backend-daily-greens/config"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type History struct {
	Id               int       `db:"id" json:"id"`
	NoInvoice        string    `db:"no_invoice" json:"noInvoice"`
	DateTransaction  time.Time `db:"date_transaction" json:"dateTransaction"`
	Status           string    `db:"status" json:"status"`
	TotalTransaction float64   `db:"total_transaction" json:"totalTransaction"`
	Image            string    `db:"image" json:"image"`
}

type HistoryDetail struct {
	Id               int            `json:"id" db:"id"`
	UserId           int            `json:"userId" db:"user_id"`
	NoInvoice        string         `json:"noInvoice" db:"no_invoice"`
	DateTransaction  time.Time      `json:"dateOrder" db:"date_transaction"`
	FullName         string         `json:"fullName" db:"full_name"`
	Email            string         `json:"email" db:"email"`
	Address          string         `json:"address" db:"address"`
	Phone            string         `json:"phone" db:"phone"`
	PaymentMethod    string         `json:"paymentMethod" db:"payment_method"`
	OrderMethod      string         `json:"orderMethod" db:"order_method"`
	Status           string         `json:"status" db:"status"`
	DeliveryFee      float64        `json:"deliveryFee" db:"delivery_fee"`
	AdminFee         float64        `json:"adminFee" db:"admin_fee"`
	Tax              float64        `json:"tax" db:"tax"`
	TotalTransaction float64        `json:"totalTransaction" db:"total_transaction"`
	HistoryItems     []HistoryItems `json:"historyItems" db:"-"`
}

type HistoryItems struct {
	Id              int     `json:"id" db:"id"`
	TransactionId   int     `json:"transactionId" db:"transaction_id"`
	ProductId       int     `json:"productId" db:"product_id"`
	ProductName     string  `json:"productName" db:"product_name"`
	ProductImage    string  `json:"productImage" db:"product_image"`
	ProductPrice    float64 `json:"productPrice" db:"product_price"`
	DiscountPercent float64 `json:"discountPercent" db:"discount_percent"`
	DiscountPrice   float64 `json:"discountPrice" db:"discount_price"`
	SizeName        string  `json:"sizeName" db:"size"`
	SizeCost        float64 `json:"sizeCost" db:"size_cost"`
	VariantName     string  `json:"variantName" db:"variant"`
	VariantCost     float64 `json:"variantCost" db:"variant_cost"`
	Amount          int     `json:"amount" db:"amount"`
	Subtotal        float64 `json:"subtotal" db:"subtotal"`
}

func GetListHistories(userId int, page int, limit int, date string, statusId int) ([]History, int, string, error) {
	histories := []History{}
	totalData := 0
	message := ""
	query := `SELECT 
				t.id,
				t.no_invoice,
				t.date_transaction,
				s.name AS status,
				t.total_transaction,
				COALESCE(MAX(pi.product_image), '') AS image
			FROM transactions t
			JOIN status s ON t.status_id = s.id
			JOIN transaction_items ti ON t.id = ti.transaction_id
			JOIN products p ON ti.product_id = p.id
			LEFT JOIN product_images pi ON p.id = pi.product_id AND pi.is_primary = true
			WHERE t.user_id = $1`

	params := []any{userId}
	paramIndex := 2

	if date != "" {
		query += fmt.Sprintf(" AND DATE(t.date_transaction) = $%d", paramIndex)
		params = append(params, date)
		paramIndex++
	}

	if statusId > 0 {
		query += fmt.Sprintf(" AND t.status_id = $%d", paramIndex)
		params = append(params, statusId)
		paramIndex++
	}

	groupClause := " GROUP BY t.id, s.name, t.no_invoice, t.date_transaction, t.total_transaction"

	// get total data
	countQuery := "SELECT COUNT(*) FROM (" + query + groupClause + ") AS sub"
	err := config.DB.QueryRow(context.Background(), countQuery, params...).Scan(&totalData)
	if err != nil {
		message = "Failed to count total transactions in database"
		return histories, totalData, message, err
	}

	query += groupClause

	offset := (page - 1) * limit
	query += fmt.Sprintf(" ORDER BY t.date_transaction DESC, t.id DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	params = append(params, limit, offset)

	rows, err := config.DB.Query(context.Background(), query, params...)
	if err != nil {
		message = "Failed to fetch transactions from database"
		return histories, totalData, message, err
	}
	defer rows.Close()

	histories, err = pgx.CollectRows(rows, pgx.RowToStructByName[History])
	if err != nil {
		message = "Failed to process transaction data from database"
		return histories, totalData, message, err
	}

	message = "Successfully retrieved transaction histories"
	return histories, totalData, message, nil
}

func GetDetailHistory(noInvoice string) (HistoryDetail, string, error) {
	historyDetail := HistoryDetail{}
	message := ""

	ctx := context.Background()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		message = "Failed to start database transaction"
		return historyDetail, message, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT 
			t.id,
			t.user_id,
			t.no_invoice,
			t.date_transaction,
			t.full_name,
			t.email,
			t.address,
			t.phone,
			pm.name AS payment_method,
			om.name AS order_method,
			s.name AS status,
			t.delivery_fee,
			t.admin_fee,
			t.tax,
			t.total_transaction
		FROM 
			transactions t
		JOIN 
			payment_methods pm ON t.payment_method_id = pm.id
		JOIN
			order_methods om ON t.order_method_id = om.id
		JOIN 
			status s ON t.status_id = s.id
		WHERE t.no_invoice = $1`, noInvoice)
	if err != nil {
		message = "Failed to fetch history from database"
		return historyDetail, message, err
	}
	defer rows.Close()

	historyDetail, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[HistoryDetail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message = "History not found"
			return historyDetail, message, err
		}

		message = "Failed to process history data"
		return historyDetail, message, err
	}

	itemRows, err := tx.Query(ctx,
		`SELECT 
			ti.id,
			ti.transaction_id,
			ti.product_id,
			ti.product_name,
			COALESCE(MAX(pi.product_image), '') AS product_image,
			ti.product_price,
			ti.discount_percent,
			ti.discount_price,
			ti.size,
			ti.size_cost,
			ti.variant,
			ti.variant_cost,
			ti.amount,
			ti.subtotal
		FROM transaction_items ti
		LEFT JOIN product_images pi ON ti.product_id = pi.product_id
		WHERE transaction_id = $1
		GROUP BY ti.id
		ORDER BY id ASC`, historyDetail.Id)
	if err != nil {
		message = "Failed to fetch ordered products from database"
		return historyDetail, message, err
	}
	defer itemRows.Close()

	historyItems, err := pgx.CollectRows(itemRows, pgx.RowToStructByName[HistoryItems])
	if err != nil {
		message = "Failed to process history items"
		return historyDetail, message, err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		message = "Failed to commit transaction"
		return historyDetail, message, err
	}

	historyDetail.HistoryItems = historyItems

	message = "Success get history detail"
	return historyDetail, message, nil
}
