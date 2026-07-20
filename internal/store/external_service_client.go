package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sw-liv-xd-item-update-microservice/internal/models"
)

// ExternalServiceClient maneja las llamadas al servicio externo de datos
type ExternalServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewExternalServiceClient crea una nueva instancia de ExternalServiceClient
func NewExternalServiceClient(baseURL string) *ExternalServiceClient {
	return &ExternalServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetItemDataBySKU obtiene los datos de un item desde el servicio externo
func (c *ExternalServiceClient) GetItemDataBySKU(ctx context.Context, sku string) (*models.ItemData, error) {
	if sku == "" {
		return nil, fmt.Errorf("SKU es requerido")
	}

	// Construir URL con el parámetro skus
	url := fmt.Sprintf("%s?skus=%s", c.baseURL, sku)
	log.Printf("🌐 Llamando al servicio externo: %s", url)

	// Crear request con contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear request: %w", err)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al llamar al servicio externo: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("servicio externo retornó status %d: %s", resp.StatusCode, string(body))
	}

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta del servicio externo: %w", err)
	}

	// El servicio retorna un array de items
	var items []models.ItemData
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error al parsear respuesta del servicio externo: %w", err)
	}

	// Verificar que retornó al menos un item
	if len(items) == 0 {
		log.Printf("⚠️  Servicio externo no retornó datos para SKU: %s", sku)
		return nil, nil
	}

	// Tomar el primer item
	itemData := &items[0]

	// Asegurar que el SKU esté presente
	if itemData.SKU == "" {
		itemData.SKU = sku
	}

	log.Printf("✅ Datos obtenidos del servicio externo: SKU=%s, ProductName=%s, Color=%s",
		itemData.SKU, itemData.ProductName, itemData.Color)

	return itemData, nil
}
