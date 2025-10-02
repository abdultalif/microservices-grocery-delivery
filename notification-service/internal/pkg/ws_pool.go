package pkg

import (
	"sync"

	"github.com/gorilla/websocket"
)

var (
	WsClient      = make(map[int]*websocket.Conn)
	WsClientMutex = sync.RWMutex{}
)

func AddWebSocketConn(userID int, conn *websocket.Conn) {
	WsClientMutex.Lock()
	defer WsClientMutex.Unlock()
	WsClient[userID] = conn
}

func GetWebSocketConn(userID int) *websocket.Conn {
	WsClientMutex.RLock()
	defer WsClientMutex.RUnlock()
	return WsClient[userID]
}

func RemoveWebSocketConn(userID int) {
	WsClientMutex.Lock()
	defer WsClientMutex.Unlock()
	delete(WsClient, userID)
}
