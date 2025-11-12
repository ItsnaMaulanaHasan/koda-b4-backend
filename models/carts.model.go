package models

type Cart struct {
	Id          int     `db:"id"`
	UserId      int     `db:"user_id"`
	ProductName string  `db:"product_name"`
	Size        string  `db:"size"`
	Variant     string  `db:"variant"`
	Amount      int     `db:"amount"`
	Subtotal    float64 `db:"subtotal"`
}

type CartRequest struct {
	Id        int     `json:"-"`
	UserId    int     `json:"-"`
	ProductId int     `json:"productId"`
	SizeId    int     `json:"sizeId"`
	VariantId int     `json:"variantId"`
	Amount    float64 `json:"amount"`
	Subtotal  float64 `json:"subtotal"`
}
