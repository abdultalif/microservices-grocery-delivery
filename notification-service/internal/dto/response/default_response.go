package response

type DefaultResponse struct {
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

func ResponseDefaultSuccess(code int, message string, data interface{}) DefaultResponse {
	return DefaultResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func ResponseDefaultError(code int, message string) DefaultResponse {
	return DefaultResponse{
		Success: false,
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

func ResponseSuccessWithPagination(code int, message string, data interface{}, page, totalData, totalPage, limit int64) DefaultResponseWithPaginations {
	return DefaultResponseWithPaginations{
		Success: true,
		Code:    code,
		Message: message,
		Pagination: &Pagination{
			Page:       page,
			TotalCount: totalData,
			PerPage:    limit,
			TotalPage:  totalPage,
		},
		Data: data,
	}
}
