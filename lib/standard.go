package lib

type Response struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Success/Error message"`
	Data    any    `json:"data,omitempty"`
}
