package models

import (
	"backend-daily-greens/config"
	"context"
	"fmt"

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
