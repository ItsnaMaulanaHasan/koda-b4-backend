package models

import "time"

type Transaction struct {
	Id        int       `db:"id"`
	NoOrder   string    `db:"no_order"`
	DateOrder time.Time `db:"date_order"`
	Status    string    `db:"status"`
	Total     float64   `db:"total_transaction"`
}

type TransactionDetail struct {
	Id               int              `json:"id" db:"id"`
	UserId           int              `json:"user_id" db:"user_id"`
	NoOrder          string           `json:"no_order" db:"no_order"`
	DateOrder        time.Time        `json:"date_order" db:"date_order"`
	FullName         string           `json:"full_name" db:"full_name"`
	Email            string           `json:"email" db:"email"`
	Address          string           `json:"address" db:"address"`
	Phone            string           `json:"phone" db:"phone"`
	PaymentMethod    string           `json:"payment_method" db:"payment_method"`
	Shipping         string           `json:"shipping" db:"shipping"`
	Status           string           `json:"status" db:"status"`
	TotalTransaction float64          `json:"total_transaction" db:"total_transaction"`
	DeliveryFee      float64          `json:"delivery_fee" db:"delivery_fee"`
	Tax              float64          `json:"tax" db:"tax"`
	OrderedProducts  []OrderedProduct `json:"ordered_products"`
}

type OrderedProduct struct {
	Id              int     `json:"id" db:"id"`
	ProductId       int     `json:"product_id" db:"product_id"`
	ProductName     string  `json:"product_name" db:"product_name"`
	ProductPrice    float64 `json:"product_price" db:"product_price"`
	DiscountPercent float64 `json:"discount_percent" db:"discount_percent"`
	Amount          int     `json:"amount" db:"amount"`
	Subtotal        float64 `json:"subtotal" db:"subtotal"`
	Size            string  `json:"size" db:"size"`
}
