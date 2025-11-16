package models

import (
	"backend-daily-greens/config"
	"context"

	"github.com/jackc/pgx/v5"
)

type PaymentMethod struct {
	Id       int     `json:"id" db:"id"`
	Name     string  `json:"name" db:"name"`
	AdminFee float64 `json:"adminFee" db:"admin_fee"`
}

type OrderMethod struct {
	Id          int     `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	DeliveryFee float64 `json:"deliveryFee" db:"delivery_fee"`
}

func GetAllPaymentMethods() ([]PaymentMethod, string, error) {
	paymentMethods := []PaymentMethod{}
	message := ""

	rows, err := config.DB.Query(
		context.Background(),
		`SELECT id, name, admin_fee 
		 FROM payment_methods 
		 ORDER BY id ASC`,
	)
	if err != nil {
		message = "Failed to fetch payment methods from database"
		return paymentMethods, message, err
	}
	defer rows.Close()

	paymentMethods, err = pgx.CollectRows(rows, pgx.RowToStructByName[PaymentMethod])
	if err != nil {
		message = "Failed to process payment methods data"
		return paymentMethods, message, err
	}

	message = "Success get all payment methods"
	return paymentMethods, message, nil
}

func GetAllOrderMethods() ([]OrderMethod, string, error) {
	orderMethods := []OrderMethod{}
	message := ""

	rows, err := config.DB.Query(
		context.Background(),
		`SELECT id, name, delivery_fee 
		 FROM order_methods 
		 ORDER BY id ASC`,
	)
	if err != nil {
		message = "Failed to fetch order methods from database"
		return orderMethods, message, err
	}
	defer rows.Close()

	orderMethods, err = pgx.CollectRows(rows, pgx.RowToStructByName[OrderMethod])
	if err != nil {
		message = "Failed to process order methods data"
		return orderMethods, message, err
	}

	message = "Success get all order methods"
	return orderMethods, message, nil
}
