package handlers

import (
	"errors"
	"net/http"
	"notification-service/internal/domain/entity"
	errs "notification-service/internal/domain/error"
	"notification-service/internal/dto/response"
	"notification-service/internal/pkg"
	"notification-service/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type NotifHandlerInterface interface {
	GetAll(c echo.Context) error
}

type NotifHandler struct {
	notifService services.NotificationServiceInterface
}

// GetAll implements NotifHandlerInterface.
func (n *NotifHandler) GetAll(c echo.Context) error {

	var (
		ctx = c.Request().Context()
		res = []response.ListResponse{}
	)

	user, ok := c.Get("user").(entity.JwtUserData)
	if !ok {
		log.Error("[NotifHandler-1] GetAll: user data not found in context")
		return c.JSON(http.StatusUnauthorized, response.ResponseDefaultError(http.StatusUnauthorized, "unauthorized"))
	}

	search := c.QueryParam("search")
	var page int64 = 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, _ = pkg.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var PerPage int64 = 10
	if PerPageStr := c.QueryParam("perPage"); PerPageStr != "" {
		PerPage, _ = pkg.StringToInt64(PerPageStr)
		if PerPage <= 0 {
			PerPage = 10
		}
	}

	status := ""
	if statusStr := c.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	orderBy := "created_at"
	if orderByStr := c.QueryParam("orderBy"); orderByStr != "" {
		orderBy = orderByStr
	}

	orderType := "desc"
	if orderTypeStr := c.QueryParam("orderType"); orderTypeStr != "" {
		orderType = orderTypeStr
	}

	isRead := false
	if isReadStr := c.QueryParam("isRead"); isReadStr != "" {
		if isReadStr == "true" {
			isRead = true
		}
	}

	reqEntity := entity.NotifyQuerySting{
		Search:    search,
		Page:      page,
		Limit:     PerPage,
		Status:    status,
		OrderBy:   orderBy,
		OrderType: orderType,
		UserID:    user.UserID,
		IsRead:    isRead,
	}

	results, totalData, totalPage, err := n.notifService.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[NotifHandler-1] GetAll: %v", err)
		if errors.Is(err, errs.ErrNotFoundNotification) {
			return c.JSON(http.StatusNotFound, response.ResponseDefaultError(http.StatusNotFound, err.Error()))
		} else {
			return c.JSON(http.StatusInternalServerError, response.ResponseDefaultError(http.StatusInternalServerError, err.Error()))
		}
	}

	for _, result := range results {
		res = append(res, response.ListResponse{
			ID:      result.ID,
			Subject: *result.Subject,
			Status:  result.Status,
			SentAt:  result.SentAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(http.StatusOK, response.ResponseSuccessWithPagination(http.StatusOK, "success", res, totalData, totalPage, page, PerPage))

}

func NewNotifHandler(notifService services.NotificationServiceInterface) NotifHandlerInterface {
	return &NotifHandler{notifService: notifService}
}
