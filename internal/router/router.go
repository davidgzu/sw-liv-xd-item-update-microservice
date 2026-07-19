package router

import (
	"sw-liv-xd-item-update-microservice/internal/handlers"

	"github.com/labstack/echo/v4"
)

// SetupRoutes configura todas las rutas de la aplicación
func SetupRoutes(e *echo.Echo, itemHandler *handlers.ItemHandler) {
	// Health check
	e.GET("/health", itemHandler.HealthCheck)

	// API v1
	v1 := e.Group("/api/v1")
	{
		// Endpoint para recibir mensajes de Pub/Sub
		v1.POST("/items/pubsub", itemHandler.HandlePubSubMessage)
	}
}
