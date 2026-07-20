package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sw-liv-xd-item-update-microservice/internal/models"

	"github.com/labstack/echo/v4"
)

// MockItemService es un mock del servicio de items
type MockItemService struct {
	processFunc func(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemisionDB, error)
}

func (m *MockItemService) ProcessItemUpdate(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemisionDB, error) {
	return m.processFunc(ctx, request)
}

func TestHealthCheck(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockItemService{}
	handler := NewItemHandler(mockService)

	// Test
	if err := handler.HealthCheck(c); err != nil {
		t.Errorf("HealthCheck retornó error: %v", err)
	}

	// Verificar status code
	if rec.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, obtenido: %d", http.StatusOK, rec.Code)
	}

	// Verificar respuesta
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Error al parsear respuesta: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Status esperado: 'healthy', obtenido: '%s'", response["status"])
	}
}

func TestHandlePubSubMessage_Success(t *testing.T) {
	// Crear mensaje de prueba
	itemRequest := models.ItemUpdateRequest{
		LogObject: models.LogObject{
			IDRemision:     123,
			IDItemRemision: 456,
			OrderNumber:    "ORD-12345",
			SKU:            "1033804373",
			ItemDesc:       "Apple iPhone 16 Plus con pantalla de 6.7 pulgadas y 256GB de almacenamiento interno",
			ItemShortDesc:  "iPhone 16 Plus",
		},
	}

	// Codificar a JSON y luego a base64
	itemRequestJSON, _ := json.Marshal(itemRequest)
	base64Data := base64.StdEncoding.EncodeToString(itemRequestJSON)

	// Crear mensaje Pub/Sub
	pubsubMessage := models.PubSubMessage{}
	pubsubMessage.Message.Data = base64Data

	// Setup Echo
	e := echo.New()
	body, _ := json.Marshal(pubsubMessage)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/pubsub", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock service
	mockService := &MockItemService{
		processFunc: func(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemisionDB, error) {
			// Verificar que los datos se parsearon correctamente
			if request.LogObject.SKU != "1033804373" {
				t.Errorf("SKU esperado: 1033804373, obtenido: %s", request.LogObject.SKU)
			}
			if request.LogObject.IDItemRemision != 456 {
				t.Errorf("IDItemRemision esperado: 456, obtenido: %d", request.LogObject.IDItemRemision)
			}

			// Retornar item actualizado
			return &models.ItemRemisionDB{
				IDItemRemision: 456,
				MaterialName:   "iPhone 16 Plus | Color: Negro | Talla: 256 GB",
				ImageURL:       sql.NullString{String: "https://example.com/image.jpg", Valid: true},
			}, nil
		},
	}

	handler := NewItemHandler(mockService)

	// Test
	if err := handler.HandlePubSubMessage(c); err != nil {
		t.Errorf("HandlePubSubMessage retornó error: %v", err)
	}

	// Verificar status code
	if rec.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, obtenido: %d", http.StatusOK, rec.Code)
	}
}

func TestHandlePubSubMessage_InvalidJSON(t *testing.T) {
	// Setup Echo con JSON inválido
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/pubsub", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockItemService{}
	handler := NewItemHandler(mockService)

	// Test
	handler.HandlePubSubMessage(c)

	// Verificar status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status code esperado: %d, obtenido: %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlePubSubMessage_InvalidBase64(t *testing.T) {
	// Crear mensaje Pub/Sub con base64 inválido
	pubsubMessage := models.PubSubMessage{}
	pubsubMessage.Message.Data = "invalid-base64!!!"

	// Setup Echo
	e := echo.New()
	body, _ := json.Marshal(pubsubMessage)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/pubsub", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockItemService{}
	handler := NewItemHandler(mockService)

	// Test
	handler.HandlePubSubMessage(c)

	// Verificar status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status code esperado: %d, obtenido: %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlePubSubMessage_EmptySKU(t *testing.T) {
	// Crear mensaje de prueba sin SKU
	itemRequest := models.ItemUpdateRequest{
		LogObject: models.LogObject{
			IDRemision:     123,
			IDItemRemision: 456,
			OrderNumber:    "ORD-12345",
			SKU:            "", // SKU vacío
			ItemDesc:       "Descripción",
			ItemShortDesc:  "Desc corta",
		},
	}

	// Codificar a JSON y luego a base64
	itemRequestJSON, _ := json.Marshal(itemRequest)
	base64Data := base64.StdEncoding.EncodeToString(itemRequestJSON)

	// Crear mensaje Pub/Sub
	pubsubMessage := models.PubSubMessage{}
	pubsubMessage.Message.Data = base64Data

	// Setup Echo
	e := echo.New()
	body, _ := json.Marshal(pubsubMessage)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/pubsub", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockItemService{}
	handler := NewItemHandler(mockService)

	// Test
	handler.HandlePubSubMessage(c)

	// Verificar status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status code esperado: %d, obtenido: %d", http.StatusBadRequest, rec.Code)
	}

	// Verificar mensaje de error
	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	if response["error"] != "SKU is required" {
		t.Errorf("Mensaje de error esperado: 'SKU is required', obtenido: '%s'", response["error"])
	}
}
