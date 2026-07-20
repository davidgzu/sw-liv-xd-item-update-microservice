package services

import (
	"context"
	"fmt"
	"log"

	"sw-liv-xd-item-update-microservice/internal/config"
	"sw-liv-xd-item-update-microservice/internal/models"
)

// FirestoreStore define la interfaz para operaciones de Firestore
type FirestoreStore interface {
	GetItemDataBySKU(ctx context.Context, sku string) (*models.ItemData, error)
	SaveItemData(ctx context.Context, itemData *models.ItemData) error
}

// ExternalService define la interfaz para el servicio externo de datos
type ExternalService interface {
	GetItemDataBySKU(ctx context.Context, sku string) (*models.ItemData, error)
}

// MySQLStore define la interfaz para operaciones de MySQL
type MySQLStore interface {
	GetItemRemisionByID(ctx context.Context, idItemRemision int64) (*models.ItemRemisionDB, error)
	UpdateItemRemision(ctx context.Context, idItemRemision int64, itemData *models.ItemData, itemDesc, itemShortDesc string) error
}

// ItemService maneja la lógica de negocio para items
type ItemService struct {
	firestoreStore  FirestoreStore
	mysqlStore      MySQLStore
	externalService ExternalService
	config          *config.AppConfig
}

// NewItemService crea una nueva instancia de ItemService
func NewItemService(firestoreStore FirestoreStore, mysqlStore MySQLStore, externalService ExternalService, cfg *config.AppConfig) *ItemService {
	return &ItemService{
		firestoreStore:  firestoreStore,
		mysqlStore:      mysqlStore,
		externalService: externalService,
		config:          cfg,
	}
}

// ProcessItemUpdate procesa la actualización de un item
func (s *ItemService) ProcessItemUpdate(ctx context.Context, request *models.ItemUpdateRequest) (*models.ItemRemisionDB, error) {
	// Obtener datos del mensaje
	sku := request.LogObject.SKU
	idItemRemision := request.LogObject.IDItemRemision
	itemDesc := request.LogObject.ItemDesc
	itemShortDesc := request.LogObject.ItemShortDesc

	log.Printf("📦 Procesando item: SKU=%s, IDItemRemision=%d",
		sku, idItemRemision)
	log.Printf("   ItemDesc: %s", itemDesc)
	log.Printf("   ItemShortDesc: %s", itemShortDesc)

	// 1. Buscar datos del item en Firestore usando SKU como document ID
	itemData, err := s.firestoreStore.GetItemDataBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("error al buscar item en Firestore: %w", err)
	}

	// Si no se encuentra en Firestore, buscar en servicio externo
	if itemData == nil {
		log.Printf("⚠️  Item con SKU %s no encontrado en Firestore, buscando en servicio externo...", sku)

		// Llamar al servicio externo
		itemData, err = s.externalService.GetItemDataBySKU(ctx, sku)
		if err != nil {
			return nil, fmt.Errorf("error al buscar item en servicio externo: %w", err)
		}

		if itemData == nil {
			return nil, fmt.Errorf("item con SKU %s no encontrado en servicio externo", sku)
		}

		log.Printf("✅ Datos obtenidos del servicio externo: ProductName=%s, Color=%s, Talla=%s",
			itemData.ProductName, itemData.Color, itemData.TamanoUnico)

		// Guardar en Firestore para futuras consultas
		if err := s.firestoreStore.SaveItemData(ctx, itemData); err != nil {
			log.Printf("⚠️  Error al guardar item en Firestore (no crítico): %v", err)
			// No retornar error, continuar con el flujo
		} else {
			log.Printf("💾 Item guardado en Firestore para futuras consultas: SKU=%s", sku)
		}
	} else {
		log.Printf("✅ Datos encontrados en Firestore: ProductName=%s, Color=%s, Talla=%s",
			itemData.ProductName, itemData.Color, itemData.TamanoUnico)
	}

	// 2. Verificar que el ItemRemision existe en MySQL
	existingItem, err := s.mysqlStore.GetItemRemisionByID(ctx, idItemRemision)
	if err != nil {
		return nil, fmt.Errorf("error al buscar ItemRemision en MySQL: %w", err)
	}

	if existingItem == nil {
		return nil, fmt.Errorf("ItemRemision con ID %d no encontrado en MySQL", idItemRemision)
	}

	// 3. Actualizar ItemRemision en MySQL con datos de Firestore y del mensaje
	err = s.mysqlStore.UpdateItemRemision(ctx, idItemRemision, itemData, itemDesc, itemShortDesc)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar ItemRemision en MySQL: %w", err)
	}

	log.Printf("✅ ItemRemision actualizado: ID=%d, SKU=%s", idItemRemision, itemData.SKU)

	// 4. Retornar el item actualizado
	updatedItem, err := s.mysqlStore.GetItemRemisionByID(ctx, idItemRemision)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ItemRemision actualizado: %w", err)
	}

	return updatedItem, nil
}
