package handlers

import (
	"net/http"
	"notification-service/internal/pkg"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebSocketHandlerInterface interface {
	WebSocketHandler(c echo.Context) error
}

type webSocketHandler struct {
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (ws *webSocketHandler) WebSocketHandler(c echo.Context) error {
	userIDStr := c.QueryParam("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user_id")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	pkg.AddWebSocketConn(userID, conn)
	defer pkg.RemoveWebSocketConn(userID)
	defer conn.Close()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
	return nil
}

func NewWebSocketHandler(e *echo.Echo) WebSocketHandlerInterface {
	return &webSocketHandler{}

}
