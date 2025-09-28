package main

import (
	httpinfra "notification-service/internal/infrastructure/http"
)

func main() {
	httpinfra.StartHTTPServer()
}
