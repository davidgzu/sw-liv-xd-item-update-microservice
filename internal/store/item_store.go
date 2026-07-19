package store

import (
	"context"
	"database/sql"
	"fmt"
	"sw-liv-xd-item-update-microservice/internal/models"

	"github.com/jmoiron/sqlx"
)

// ItemStore maneja las operaciones de base de datos para items
type ItemStore struct {
	db *sqlx.DB
}

// NewItemStore crea una nueva instancia de ItemStore
func NewItemStore(db *sqlx.DB) *ItemStore {
	return &ItemStore{
		db: db,
	}
}

// GetItemBySKU busca un item por SKU en la tabla item_remision
func (s *ItemStore) GetItemBySKU(ctx context.Context, sku string) (*models.ItemRemision, error) {
	query := `
		SELECT id, sku, name, description, price, stock, status, created_at, updated_at
		FROM item_remision
		WHERE sku = ?
		LIMIT 1
	`

	var item models.ItemRemision
	err := s.db.GetContext(ctx, &item, query, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar item por SKU: %w", err)
	}

	return &item, nil
}

// CreateItem crea un nuevo item en la tabla item_remision
func (s *ItemStore) CreateItem(ctx context.Context, item *models.ItemRemision) (int64, error) {
	query := `
		INSERT INTO item_remision (sku, name, description, price, stock, status, created_at, updated_at)
		VALUES (:sku, :name, :description, :price, :stock, :status, NOW(), NOW())
	`

	result, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return 0, fmt.Errorf("error al crear item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID del item creado: %w", err)
	}

	return id, nil
}

// UpdateItem actualiza un item existente en la tabla item_remision
func (s *ItemStore) UpdateItem(ctx context.Context, item *models.ItemRemision) error {
	query := `
		UPDATE item_remision
		SET name = :name,
		    description = :description,
		    price = :price,
		    stock = :stock,
		    status = :status,
		    updated_at = NOW()
		WHERE id = :id
	`

	result, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("error al actualizar item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el item con ID %d", item.ID)
	}

	return nil
}

// UpdateItemBySKU actualiza un item por SKU
func (s *ItemStore) UpdateItemBySKU(ctx context.Context, item *models.ItemRemision) error {
	query := `
		UPDATE item_remision
		SET name = :name,
		    description = :description,
		    price = :price,
		    stock = :stock,
		    status = :status,
		    updated_at = NOW()
		WHERE sku = :sku
	`

	result, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("error al actualizar item por SKU: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el item con SKU %s", item.SKU)
	}

	return nil
}
