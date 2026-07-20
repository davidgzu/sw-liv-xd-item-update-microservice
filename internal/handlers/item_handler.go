package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"sw-liv-xd-item-update-microservice/internal/models"

	"github.com/labstack/echo/v4"
)

// ItemService define la interfaz del servicio de items
type ItemService interface {
	ProcessItemUpdate(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemisionDB, error)
}

// ItemHandler maneja las peticiones HTTP relacionadas con items
type ItemHandler struct {
	service ItemService
}

// NewItemHandler crea una nueva instancia de ItemHandler
func NewItemHandler(service ItemService) *ItemHandler {
	return &ItemHandler{
		service: service,
	}
}

// HealthCheck maneja el health check del servicio
func (h *ItemHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// HandlePubSubMessage maneja los mensajes recibidos de Pub/Sub
func (h *ItemHandler) HandlePubSubMessage(c echo.Context) error {
	var pubsubMsg models.PubSubMessage

	// Parsear el mensaje de Pub/Sub
	if err := c.Bind(&pubsubMsg); err != nil {
		log.Printf("Error al parsear mensaje de Pub/Sub: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid Pub/Sub message format",
		})
	}

	// Decodificar el payload base64
	decodedData, err := base64.StdEncoding.DecodeString(pubsubMsg.Message.Data)
	if err != nil {
		log.Printf("Error al decodificar mensaje base64: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to decode message data",
		})
	}

	// Parsear el mensaje de item
	var itemRequest models.ItemUpdateRequest
	if err := json.Unmarshal(decodedData, &itemRequest); err != nil {
		log.Printf("Error al parsear item request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid item request format",
		})
	}

	// Validar que venga el SKU
	if itemRequest.SKU == "" {
		log.Printf("SKU vacío en la petición")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "SKU is required",
		})
	}

	log.Printf("Procesando item con SKU: %s (idRemision: %d, idItemRemision: %d, orderNumber: %s)",
		itemRequest.SKU,
		itemRequest.IDRemision,
		itemRequest.IDItemRemision,
		itemRequest.OrderNumber)

	// Procesar el item
	ctx := c.Request().Context()
	result, err := h.service.ProcessItemUpdate(ctx, &itemRequest)
	if err != nil {
		// Si es un error de "SKU sin datos", retornar 200 para evitar reintentos
		if models.IsNoDataError(err) {
			log.Printf("⚠️  Proceso terminado: %v (no se reintentará)", err)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success": false,
				"message": "SKU sin datos disponibles",
				"reason":  err.Error(),
				"sku":     itemRequest.SKU,
			})
		}
		// Para otros errores, retornar 500 para que Pub/Sub reintente
		log.Printf("❌ Error al procesar item: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process item",
		})
	}

	log.Printf("✅ Item procesado exitosamente: %s", itemRequest.SKU)

	// Responder con éxito (Pub/Sub espera 200-299 para ACK)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Item processed successfully",
		"data":    result,
	})
}
