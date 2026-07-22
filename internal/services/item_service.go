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
	UpdateItemRemision(ctx context.Context, idItemRemision int64, itemData *models.ItemData) error
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
func (s *ItemService) ProcessItemUpdate(ctx context.Context, request *models.ItemUpdateRequest) error {
	// Obtener datos del mensaje
	sku := request.SKU
	idItemRemision := request.IDItemRemision

	log.Printf("📦 Procesando item: SKU=%s, IDItemRemision=%d", sku, idItemRemision)

	// 1. Buscar datos del item en Firestore usando SKU como document ID
	itemData, err := s.firestoreStore.GetItemDataBySKU(ctx, sku)

	// Si Firestore retorna error o no encuentra el documento, usar servicio externo
	if err != nil || itemData == nil {
		if err != nil {
			log.Printf("⚠️  Error al buscar en Firestore: %v", err)
			log.Printf("⚠️  Intentando fallback al servicio externo...")
		} else {
			log.Printf("⚠️  Item con SKU %s no encontrado en Firestore, buscando en servicio externo...", sku)
		}

		// Llamar al servicio externo
		itemData, err = s.externalService.GetItemDataBySKU(ctx, sku)
		if err != nil {
			return fmt.Errorf("error al buscar item en servicio externo: %w", err)
		}

		if itemData == nil {
			// SKU no encontrado en servicio externo - guardar en Firestore como "sin datos"
			log.Printf("❌ Item con SKU %s no encontrado en servicio externo, guardando en Firestore con HasData=false", sku)
			emptyItem := &models.ItemData{
				SKU:     sku,
				HasData: false,
			}
			// Guardar en Firestore para evitar búsquedas futuras (best effort)
			if err := s.firestoreStore.SaveItemData(ctx, emptyItem); err != nil {
				log.Printf("⚠️  Error al guardar SKU vacío en Firestore (no crítico): %v", err)
			} else {
				log.Printf("💾 SKU guardado en Firestore como 'sin datos': %s", sku)
			}
			return fmt.Errorf("%w: SKU %s no encontrado", models.ErrSKUNoData, sku)
		}

		log.Printf("✅ Datos obtenidos del servicio externo: ProductName=%s, Color=%s, Talla=%s",
			itemData.ProductName, itemData.Color, itemData.TamanoUnico)

		// Marcar como válido y guardar en Firestore
		itemData.HasData = true
		if err := s.firestoreStore.SaveItemData(ctx, itemData); err != nil {
			log.Printf("⚠️  Error al guardar item en Firestore (no crítico): %v", err)
		} else {
			log.Printf("💾 Item guardado en Firestore para futuras consultas: SKU=%s", sku)
		}
	} else {
		// Encontrado en Firestore - verificar si tiene datos válidos
		if !itemData.HasData {
			log.Printf("🚫 SKU %s marcado en Firestore como 'sin datos' (HasData=false), terminando proceso sin reintentos", sku)
			return fmt.Errorf("%w: SKU %s previamente verificado", models.ErrSKUNoData, sku)
		}
		log.Printf("✅ Datos encontrados en Firestore: ProductName=%s, Color=%s, Talla=%s",
			itemData.ProductName, itemData.Color, itemData.TamanoUnico)
	}

	// 2. Validar que ProductName tenga datos (campo crítico)
	if itemData.ProductName == "" {
		log.Printf("❌ SKU %s tiene ProductName vacío, no se puede generar descripción válida", sku)
		log.Printf("🚫 Terminando proceso sin actualizar MySQL (no se reintentará)")
		return fmt.Errorf("%w: SKU %s con ProductName vacío", models.ErrSKUNoData, sku)
	}

	// Advertir si Color o TamañoUnico están vacíos (no crítico, se continúa)
	if itemData.Color == "" {
		log.Printf("⚠️  SKU %s no tiene Color, se generará descripción sin este campo", sku)
	}
	if itemData.TamanoUnico == "" {
		log.Printf("⚠️  SKU %s no tiene TamañoUnico, se generará descripción sin este campo", sku)
	}

	// 3. Actualizar ItemRemision en MySQL con datos de Firestore/servicio externo
	// (UpdateItemRemision valida que el registro exista verificando rowsAffected)
	err = s.mysqlStore.UpdateItemRemision(ctx, idItemRemision, itemData)
	if err != nil {
		return fmt.Errorf("error al actualizar ItemRemision en MySQL: %w", err)
	}

	log.Printf("✅ ItemRemision actualizado exitosamente: ID=%d, SKU=%s", idItemRemision, itemData.SKU)

	return nil
}
