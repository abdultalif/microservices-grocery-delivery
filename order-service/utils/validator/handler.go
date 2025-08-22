package validator

import (
	"net/http"
	"order-service/internal/adapter/handler/response"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func HandleValidationError(c echo.Context, err error, trans ut.Translator) error {
	res := response.ResponseDefault{
		Success: false,
		Code:    http.StatusBadRequest,
		Data:    nil,
	}

	switch e := err.(type) {
	case validator.ValidationErrors:
		errMap := map[string][]string{}
		for _, fieldErr := range e {
			field := strings.ToLower(fieldErr.Field())
			msg := fieldErr.Translate(trans) 
			errMap[field] = append(errMap[field], msg)
		}
		res.Message = errMap

	default:
		res.Message = err.Error()
	}

	return c.JSON(http.StatusBadRequest, res)
}