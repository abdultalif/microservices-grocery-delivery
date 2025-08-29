package response

type ResponseDefault struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

type DefaultResponseWithPaginations struct {
	Success    bool        `json:"success"`
	Code       int         `json:"code"`
	Message    interface{} `json:"message"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Data       interface{} `json:"data"`
}

type Pagination struct {
	Page       int64 `json:"page"`
	TotalCount int64 `json:"total_count"`
	PerPage    int64 `json:"per_page"`
	TotalPage  int64 `json:"total_page"`
}

func APIResponseError(code int, message interface{}) ResponseDefault {
	return ResponseDefault{
		Success: false,
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

func APIResponseSuccess(code int, message, data interface{}) ResponseDefault {
	return ResponseDefault{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
}
