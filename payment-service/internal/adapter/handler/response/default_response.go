package response

type DefaultResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

func ResponseAPISuccess(code int, message interface{}, data interface{}) DefaultResponse {
	return DefaultResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func ResponseAPIError(code int, message interface{}) DefaultResponse {
	return DefaultResponse{
		Success: false,
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
