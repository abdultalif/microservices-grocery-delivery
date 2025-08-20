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

func ResponseAPI(success bool, code int, message interface{}, data interface{}) ResponseDefault {
	return ResponseDefault{
		Success: success,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func ResponseAPIWithPagination(success bool, code int, message interface{}, data interface{}, page, totalData, totalPage, limit int64) DefaultResponseWithPaginations {
	return DefaultResponseWithPaginations{
		Success: success,
		Code:    code,
		Message: message,
		Pagination: &Pagination{
			Page:       page,
			PerPage:    limit,
			TotalCount: totalData,
			TotalPage:  totalPage,
		},
		Data: data,
	}
}