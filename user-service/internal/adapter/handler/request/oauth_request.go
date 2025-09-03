package request

type LinkAccountRequest struct {
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
	Provider string `json:"provider" validate:"required"`
}
