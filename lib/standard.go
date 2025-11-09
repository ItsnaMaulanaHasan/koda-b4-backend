package lib

type ResponseSuccess struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Success message"`
	Data    any    `json:"data,omitempty"`
}

type ResponseError struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error message"`
	Error   string `json:"error,omitempty" example:"cause of error"`
}
