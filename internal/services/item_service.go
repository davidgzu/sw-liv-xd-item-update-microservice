package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sw-liv-xd-item-update-microservice/internal/config"
	"sw-liv-xd-item-update-microservice/internal/models"
)

// ItemStore define la interfaz para operaciones de items en Firestore
type ItemStore interface {
	GetItemBySKU(ctx context.Context, sku string) (*models.ItemRemision, error)
	CreateItem(ctx context.Context, item *models.ItemRemision) error
	UpdateItem(ctx context.Context, item *models.ItemRemision) error
}

// ItemService maneja la lógica de negocio para items
type ItemService struct {
	store      ItemStore
	httpClient *http.Client
	config     *config.AppConfig
}

// NewItemService crea una nueva instancia de ItemService
func NewItemService(store ItemStore, cfg *config.AppConfig) *ItemService {
	return &ItemService{
		store: store,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
	}
}

// ProcessItemUpdate procesa la actualización de un item
func (s *ItemService) ProcessItemUpdate(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemision, error) {
	// Obtener SKU del mensaje
	sku := request.LogObject.SKU

	// 1. Buscar el item en Firestore usando SKU como document ID
	item, err := s.store.GetItemBySKU(ctx, sku)

	if err != nil {
		return nil, fmt.Errorf("error al buscar item en Firestore: %w", err)
	}

	if item != nil {
		// Item encontrado en Firestore - retornar directamente
		return item, nil
	}

	// 2. Item no encontrado en Firestore - buscar en servicio externo
	externalItem, err := s.fetchFromExternalService(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("error al consultar servicio externo: %w", err)
	}

	// 3. Crear nuevo item con datos del servicio externo
	newItem := &models.ItemRemision{
		SKU:         externalItem.SKU,
		Name:        externalItem.Name,
		Description: externalItem.Description,
		Price:       externalItem.Price,
		Stock:       externalItem.Stock,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 4. Guardar en Firestore usando SKU como document ID
	err = s.store.CreateItem(ctx, newItem)
	if err != nil {
		return nil, fmt.Errorf("error al guardar item en Firestore: %w", err)
	}

	return newItem, nil
}

// fetchFromExternalService consulta el servicio externo
func (s *ItemService) fetchFromExternalService(ctx context.Context, sku string) (*models.ExternalServiceResponse, error) {
	url := fmt.Sprintf("%s/api/items/sku/%s", s.config.External.ServiceURL, sku)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear petición: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar petición: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servicio externo respondió con código: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta: %w", err)
	}

	var externalResp models.ExternalServiceResponse
	if err := json.Unmarshal(body, &externalResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	if !externalResp.Success {
		return nil, fmt.Errorf("servicio externo no encontró el item: %s", externalResp.Message)
	}

	return &externalResp, nil
}
