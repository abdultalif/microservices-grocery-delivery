package main

import (
	httpinfra "github.com/abdultalif/microservices-grocery-delivery/notification-service/internal/infrastructure/http"
)

func main() {
	httpinfra.StartHTTPServer()
}
