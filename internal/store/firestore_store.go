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
func (s *FirestoreStore) GetItemBySKU(ctx context.Context, sku string) (*models.ItemRemision, error) {
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

	// Convertir los datos a ItemRemision
	var item models.ItemRemision
	if err := docSnap.DataTo(&item); err != nil {
		return nil, fmt.Errorf("error al deserializar documento de Firestore: %w", err)
	}

	// Asegurar que el SKU esté presente (por si no viene en el documento)
	if item.SKU == "" {
		item.SKU = sku
	}

	log.Printf("✅ Item encontrado en Firestore: SKU=%s, Name=%s", item.SKU, item.Name)
	return &item, nil
}

// CreateItem guarda un item en Firestore usando el SKU como document ID
func (s *FirestoreStore) CreateItem(ctx context.Context, item *models.ItemRemision) error {
	log.Printf("💾 Guardando item en Firestore con SKU: %s", item.SKU)

	docRef := s.client.Collection(s.collection).Doc(item.SKU)
	_, err := docRef.Set(ctx, item)
	if err != nil {
		return fmt.Errorf("error al guardar item en Firestore: %w", err)
	}

	log.Printf("✅ Item guardado en Firestore: %s", item.SKU)
	return nil
}

// UpdateItem actualiza un item en Firestore
func (s *FirestoreStore) UpdateItem(ctx context.Context, item *models.ItemRemision) error {
	log.Printf("🔄 Actualizando item en Firestore: %s", item.SKU)

	docRef := s.client.Collection(s.collection).Doc(item.SKU)
	_, err := docRef.Set(ctx, item, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("error al actualizar item en Firestore: %w", err)
	}

	log.Printf("✅ Item actualizado en Firestore: %s", item.SKU)
	return nil
}
