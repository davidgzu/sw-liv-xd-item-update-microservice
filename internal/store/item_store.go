package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sw-liv-xd-item-update-microservice/internal/models"
	"time"

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

// GetItemRemisionByID busca un ItemRemision por IDItemRemision
func (s *ItemStore) GetItemRemisionByID(ctx context.Context, idItemRemision int64) (*models.ItemRemisionDB, error) {
	query := `
		SELECT * FROM ItemRemision
		WHERE IDItemRemision = ?
		LIMIT 1
	`

	var item models.ItemRemisionDB
	err := s.db.GetContext(ctx, &item, query, idItemRemision)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar ItemRemision por ID: %w", err)
	}

	return &item, nil
}

// UpdateItemRemision actualiza los campos de un ItemRemision con datos de Firestore
func (s *ItemStore) UpdateItemRemision(ctx context.Context, idItemRemision int64, itemData *models.ItemData, itemDesc, itemShortDesc string) error {
	// Generar material_name con lógica de prioridades
	materialName := s.buildMaterialName(itemData, itemDesc, itemShortDesc)

	log.Printf("🔍 Validando imageURL: %s", itemData.ImageURL)

	// Validar si el imageURL es válido
	isValidURL := s.isValidImageURL(itemData.ImageURL)

	var query string
	var args []interface{}

	if isValidURL {
		// URL válida: actualizar material_name e image_url
		log.Printf("✅ URL válida, actualizando material_name e image_url")
		query = `
			UPDATE ItemRemision
			SET
				material_name = ?,
				image_url = ?
			WHERE IDItemRemision = ?
		`
		args = []interface{}{
			materialName,
			itemData.ImageURL,
			idItemRemision,
		}
	} else {
		// URL inválida: actualizar solo material_name
		log.Printf("⚠️  URL inválida o vacía, actualizando solo material_name")
		query = `
			UPDATE ItemRemision
			SET
				material_name = ?
			WHERE IDItemRemision = ?
		`
		args = []interface{}{
			materialName,
			idItemRemision,
		}
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error al actualizar ItemRemision: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró ItemRemision con ID %d", idItemRemision)
	}

	return nil
}

// isValidImageURL valida si una URL de imagen es válida y accesible
func (s *ItemStore) isValidImageURL(imageURL string) bool {
	// Si está vacía, no es válida
	if imageURL == "" {
		return false
	}

	// Validar que sea una URL HTTP/HTTPS
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		log.Printf("❌ URL no comienza con http:// o https://")
		return false
	}

	// Hacer una petición HEAD para verificar que la URL existe
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // No seguir redirects
		},
	}

	req, err := http.NewRequest("HEAD", imageURL, nil)
	if err != nil {
		log.Printf("❌ Error al crear request: %v", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Error al hacer petición HEAD: %v", err)
		return false
	}
	defer resp.Body.Close()

	// Considerar válida si el código de respuesta es 2xx o 3xx
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		log.Printf("✅ URL válida (Status: %d)", resp.StatusCode)
		return true
	}

	log.Printf("❌ URL retorna status code: %d", resp.StatusCode)
	return false
}

// buildMaterialName construye la descripción del material con lógica de prioridades
func (s *ItemStore) buildMaterialName(itemData *models.ItemData, itemDesc, itemShortDesc string) string {
	var baseName string

	// Prioridad 1: ItemDesc si tiene contenido significativo (>50 caracteres)
	if len(itemDesc) > 50 {
		baseName = itemDesc
		// Verificar si itemDesc ya contiene color y talla
		if s.containsColorAndSize(itemDesc, itemData.Color, itemData.TamanoUnico) {
			return itemDesc // Ya tiene todo, retornar tal cual
		}
	} else if itemShortDesc != "" {
		// Prioridad 2: ItemShortDesc si existe
		baseName = itemShortDesc
		// Verificar si itemShortDesc ya contiene color y talla
		if s.containsColorAndSize(itemShortDesc, itemData.Color, itemData.TamanoUnico) {
			return itemShortDesc // Ya tiene todo, retornar tal cual
		}
	} else {
		// Prioridad 3: ProductName de Firestore
		baseName = itemData.ProductName
	}

	// Agregar color y talla si existen y no están en el texto base
	var parts []string
	if baseName != "" {
		parts = append(parts, baseName)
	}

	if itemData.Color != "" {
		parts = append(parts, "Color: "+itemData.Color)
	}

	if itemData.TamanoUnico != "" {
		parts = append(parts, "Talla: "+itemData.TamanoUnico)
	}

	// Unir las partes con " | "
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " | "
		}
		result += part
	}

	return result
}

// containsColorAndSize verifica si el texto ya contiene información de color y talla
func (s *ItemStore) containsColorAndSize(text, color, size string) bool {
	if color == "" && size == "" {
		return false // No hay nada que validar
	}

	hasColor := s.containsIgnoreCase(text, color)
	hasSize := s.containsIgnoreCase(text, size)

	// Si ambos campos existen en los datos, verificar que ambos estén en el texto
	if color != "" && size != "" {
		return hasColor && hasSize
	}

	// Si solo hay color, verificar que el color esté en el texto
	if color != "" {
		return hasColor
	}

	// Si solo hay talla, verificar que la talla esté en el texto
	if size != "" {
		return hasSize
	}

	return false
}

// containsIgnoreCase verifica si una cadena contiene otra (case insensitive)
func (s *ItemStore) containsIgnoreCase(text, substr string) bool {
	if substr == "" {
		return false
	}

	textLower := strings.ToLower(text)
	substrLower := strings.ToLower(substr)

	return strings.Contains(textLower, substrLower)
}
