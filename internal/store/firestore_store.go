package store

import (
	"context"
	"fmt"
	"log"

	"sw-liv-xd-item-update-microservice/internal/models"

	"cloud.google.com/go/firestore"
)

// FirestoreStore maneja las operaciones de Firestore para items
type FirestoreStore struct {
	client     *firestore.Client
	collection string
}

// NewFirestoreStore crea una nueva instancia de FirestoreStore
func NewFirestoreStore(client *firestore.Client, collectionName string) *FirestoreStore {
	return &FirestoreStore{
		client:     client,
		collection: collectionName,
	}
}

// GetItemBySKU busca un item en Firestore usando el SKU como document ID
func (s *FirestoreStore) GetItemDataBySKU(ctx context.Context, sku string) (*models.ItemData, error) {
	log.Printf("🔍 Buscando en Firestore colección '%s' documento ID: %s", s.collection, sku)

	// Obtener el documento usando el SKU como ID
	docRef := s.client.Collection(s.collection).Doc(sku)
	docSnap, err := docRef.Get(ctx)

	if err != nil {
		// Verificar si el documento no existe
		if err.Error() == "rpc error: code = NotFound desc = Document not found" {
			log.Printf("❌ Item con SKU %s no encontrado en Firestore", sku)
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar item en Firestore: %w", err)
	}

	if !docSnap.Exists() {
		log.Printf("❌ Documento %s no existe en Firestore", sku)
		return nil, nil
	}

	// Convertir los datos a ItemData
	var itemData models.ItemData
	if err := docSnap.DataTo(&itemData); err != nil {
		return nil, fmt.Errorf("error al deserializar documento de Firestore: %w", err)
	}

	// Asegurar que el SKU esté presente (por si no viene en el documento)
	if itemData.SKU == "" {
		itemData.SKU = sku
	}

	log.Printf("✅ Item encontrado en Firestore: SKU=%s, ProductName=%s, Color=%s",
		itemData.SKU, itemData.ProductName, itemData.Color)
	return &itemData, nil
}

// SaveItemData guarda o actualiza un item en Firestore usando el SKU como document ID
func (s *FirestoreStore) SaveItemData(ctx context.Context, itemData *models.ItemData) error {
	if itemData.SKU == "" {
		return fmt.Errorf("SKU es requerido para guardar en Firestore")
	}

	log.Printf("💾 Guardando en Firestore colección '%s' documento ID: %s", s.collection, itemData.SKU)

	// Guardar usando el SKU como ID del documento
	docRef := s.client.Collection(s.collection).Doc(itemData.SKU)
	_, err := docRef.Set(ctx, itemData)

	if err != nil {
		return fmt.Errorf("error al guardar item en Firestore: %w", err)
	}

	log.Printf("✅ Item guardado en Firestore: SKU=%s, ProductName=%s",
		itemData.SKU, itemData.ProductName)
	return nil
}
