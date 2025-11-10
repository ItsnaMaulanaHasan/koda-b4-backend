package models

import "time"

type Transaction struct {
	Id        int       `db:"id"`
	NoOrder   string    `db:"no_order"`
	DateOrder time.Time `db:"date_order"`
	Status    string    `db:"status"`
	Total     float64   `db:"total_transaction"`
}
