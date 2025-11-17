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

type HateoasLink struct {
	Self any `json:"self,omitempty"`
	Next any `json:"next,omitempty"`
	Prev any `json:"prev,omitempty"`
	Last any `json:"last,omitempty"`
}
