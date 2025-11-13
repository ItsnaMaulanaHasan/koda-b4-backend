package models

type History struct {
	Id               int     `db:"id" json:"id"`
	NoInvoice        string  `db:"no_invoice" json:"no_invoice"`
	DateTransaction  string  `db:"date_transaction" json:"date_transaction"`
	Status           string  `db:"status" json:"status"`
	TotalTransaction float64 `db:"total_transaction" json:"total_transaction"`
	Image            string  `db:"image" json:"image"`
}
