package response

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

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
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
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

func RespondWithError(c echo.Context, code int, context string, err error) error {
	log.Errorf("%s: %v", context, err)
	resp := ResponseDefault{
		Message: err.Error(),
		Data:    nil,
	}
	return c.JSON(code, resp)
}
