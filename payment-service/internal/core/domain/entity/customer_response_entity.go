package entity

type InternalTokenResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type UserHttpClientResponse struct {
	Success bool                   `json:"success"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    CustomerResponseEntity `json:"data"`
}

type CustomerResponseEntity struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}
