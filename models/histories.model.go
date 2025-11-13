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
	Id               int     `db:"id" json:"id"`
	NoInvoice        string  `db:"no_invoice" json:"no_invoice"`
	DateTransaction  string  `db:"date_transaction" json:"date_transaction"`
	Status           string  `db:"status" json:"status"`
	TotalTransaction float64 `db:"total_transaction" json:"total_transaction"`
	Image            string  `db:"image" json:"image"`
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
	PaymentMethod    string         `json:"payment_method" db:"payment_method"`
	OrderMethod      string         `json:"orderMethod" db:"order_method"`
	Status           string         `json:"status" db:"status"`
	DeliveryFee      float64        `json:"delivery_fee" db:"delivery_fee"`
	AdminFee         float64        `json:"adminFee" db:"admin_fee"`
	Tax              float64        `json:"tax" db:"tax"`
	TotalTransaction float64        `json:"totalTransaction" db:"total_transaction"`
	HistoryItems     []HistoryItems `json:"historyItems" db:"-"`
}

type HistoryItems struct {
	Id              int     `json:"id" db:"id"`
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

func GetListHistories(userId int, page int, limit int, month int, statusId int) ([]History, int, string, error) {
	histories := []History{}
	totalData := 0
	message := ""
	query := `SELECT 
				t.id,
				t.no_invoice,
				t.date_transaction,
				s.name AS status,
				t.total_transaction,
				pi.image
			FROM transactions t
			JOIN status s ON t.status_id = s.id
			JOIN transaction_items ti ON t.id = ti.transaction_id
			JOIN products p ON ti.product_id = p.id
			JOIN product_images pi ON p.id = pi.product_id AND pi.is_primary = true
			WHERE t.user_id = $1`

	params := []any{userId}
	paramIndex := 2

	if month > 0 && month <= 12 {
		query += fmt.Sprintf(" AND EXTRACT(MONTH FROM t.date_transaction) = $%d", paramIndex)
		params = append(params, month)
		paramIndex++
	}
	if statusId > 0 {
		query += fmt.Sprintf(" AND t.status_id = $%d", paramIndex)
		params = append(params, statusId)
		paramIndex++
	}

	// get total data
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	err := config.DB.QueryRow(context.Background(), countQuery, params...).Scan(&totalData)
	if err != nil {
		message = "Failed to count total transactions in database"
		return histories, totalData, message, err
	}

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

func GetDetailHistory(userId int) (HistoryDetail, string, error) {
	historyDetail := HistoryDetail{}
	message := ""
	rows, err := config.DB.Query(context.Background(),
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
			t.total_transaction,
		FROM 
			transactions t
		JOIN 
			payment_methods pm ON t.payment_method_id = pm.id
		JOIN
			order_methods om ON t.order_method_id = pm.id
		JOIN 
			status s ON t.status_id = s.id
		WHERE t.id = $1`, userId)
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

	itemRows, err := config.DB.Query(context.Background(),
		`SELECT 
			id,
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
			subtotal
		FROM transaction_items
		WHERE transaction_id = $1
		ORDER BY id ASC`, userId)
	if err != nil {
		message = "Failed to fetch ordered products from database"
		return historyDetail, message, err
	}
	defer itemRows.Close()

	historyItems, err := pgx.CollectRows(itemRows, pgx.RowToStructByName[HistoryItems])
	if err != nil {
		message = "Failed to process ordered products data"
		return historyDetail, message, err
	}

	historyDetail.HistoryItems = historyItems

	message = "Success get history detail"
	return historyDetail, message, nil
}
