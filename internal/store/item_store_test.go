package store

import (
	"testing"

	"sw-liv-xd-item-update-microservice/internal/models"
)

func TestBuildMaterialName(t *testing.T) {
	store := &ItemStore{}

	tests := []struct {
		name          string
		itemData      *models.ItemData
		itemDesc      string
		itemShortDesc string
		expected      string
		description   string
	}{
		{
			name: "ItemDesc largo con color y talla incluidos",
			itemData: &models.ItemData{
				ProductName: "Tenis Nike",
				Color:       "Negro",
				TamanoUnico: "27",
			},
			itemDesc:      "Zapato deportivo negro con suela antideslizante talla 27 para hombre",
			itemShortDesc: "",
			expected:      "Zapato deportivo negro con suela antideslizante talla 27 para hombre",
			description:   "No agrega color/talla si ya están en itemDesc",
		},
		{
			name: "ItemDesc largo sin color/talla",
			itemData: &models.ItemData{
				ProductName: "Tenis Nike",
				Color:       "Negro",
				TamanoUnico: "27",
			},
			itemDesc:      "Zapato deportivo con suela antideslizante para hombre de alta calidad",
			itemShortDesc: "",
			expected:      "Zapato deportivo con suela antideslizante para hombre de alta calidad | Color: Negro | Talla: 27",
			description:   "Agrega color y talla al final",
		},
		{
			name: "ItemDesc corto, usa ItemShortDesc",
			itemData: &models.ItemData{
				ProductName: "Tenis Nike",
				Color:       "Negro",
				TamanoUnico: "27",
			},
			itemDesc:      "Tenis",
			itemShortDesc: "Zapato deportivo",
			expected:      "Zapato deportivo | Color: Negro | Talla: 27",
			description:   "Usa ItemShortDesc cuando ItemDesc es corto",
		},
		{
			name: "ItemShortDesc con color incluido",
			itemData: &models.ItemData{
				ProductName: "Tenis Nike",
				Color:       "Negro",
				TamanoUnico: "27",
			},
			itemDesc:      "Tenis",
			itemShortDesc: "Zapato negro deportivo talla 27",
			expected:      "Zapato negro deportivo talla 27",
			description:   "No agrega nada si color y talla ya están",
		},
		{
			name: "Sin ItemDesc ni ItemShortDesc, usa ProductName",
			itemData: &models.ItemData{
				ProductName: "Apple iPhone 16 Plus",
				Color:       "Negro",
				TamanoUnico: "256 GB",
			},
			itemDesc:      "",
			itemShortDesc: "",
			expected:      "Apple iPhone 16 Plus | Color: Negro | Talla: 256 GB",
			description:   "Usa ProductName como fallback",
		},
		{
			name: "Sin color ni talla",
			itemData: &models.ItemData{
				ProductName: "Mouse inalámbrico",
				Color:       "",
				TamanoUnico: "",
			},
			itemDesc:      "Mouse inalámbrico ergonómico con batería recargable de larga duración",
			itemShortDesc: "",
			expected:      "Mouse inalámbrico ergonómico con batería recargable de larga duración",
			description:   "Solo retorna la descripción base",
		},
		{
			name: "Solo con color, sin talla",
			itemData: &models.ItemData{
				ProductName: "Camisa",
				Color:       "Azul",
				TamanoUnico: "",
			},
			itemDesc:      "Camisa de algodón con cuello redondo y manga larga para hombre",
			itemShortDesc: "",
			expected:      "Camisa de algodón con cuello redondo y manga larga para hombre | Color: Azul",
			description:   "Solo agrega color cuando no hay talla",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.buildMaterialName(tt.itemData, tt.itemDesc, tt.itemShortDesc)
			if result != tt.expected {
				t.Errorf("\n%s\nEsperado: %s\nObtenido: %s", tt.description, tt.expected, result)
			}
		})
	}
}

func TestContainsColorAndSize(t *testing.T) {
	store := &ItemStore{}

	tests := []struct {
		name     string
		text     string
		color    string
		size     string
		expected bool
	}{
		{
			name:     "Contiene ambos",
			text:     "Zapato negro talla 27",
			color:    "Negro",
			size:     "27",
			expected: true,
		},
		{
			name:     "Contiene color case insensitive",
			text:     "Zapato NEGRO deportivo",
			color:    "negro",
			size:     "",
			expected: true,
		},
		{
			name:     "No contiene color",
			text:     "Zapato deportivo talla 27",
			color:    "Negro",
			size:     "27",
			expected: false,
		},
		{
			name:     "No contiene talla",
			text:     "Zapato negro deportivo",
			color:    "Negro",
			size:     "27",
			expected: false,
		},
		{
			name:     "Sin color ni talla en datos",
			text:     "Zapato deportivo",
			color:    "",
			size:     "",
			expected: false,
		},
		{
			name:     "Contiene solo color (sin talla en datos)",
			text:     "iPhone negro 256GB",
			color:    "Negro",
			size:     "",
			expected: true,
		},
		{
			name:     "Contiene solo talla (sin color en datos)",
			text:     "iPhone 256 GB",
			color:    "",
			size:     "256 GB",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.containsColorAndSize(tt.text, tt.color, tt.size)
			if result != tt.expected {
				t.Errorf("Esperado: %v, Obtenido: %v", tt.expected, result)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	store := &ItemStore{}

	tests := []struct {
		name     string
		text     string
		substr   string
		expected bool
	}{
		{
			name:     "Encuentra substring",
			text:     "Zapato deportivo Negro",
			substr:   "negro",
			expected: true,
		},
		{
			name:     "Case insensitive",
			text:     "iPhone NEGRO",
			substr:   "Negro",
			expected: true,
		},
		{
			name:     "No encuentra",
			text:     "Zapato deportivo",
			substr:   "negro",
			expected: false,
		},
		{
			name:     "Substring vacío",
			text:     "Zapato deportivo",
			substr:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.containsIgnoreCase(tt.text, tt.substr)
			if result != tt.expected {
				t.Errorf("Esperado: %v, Obtenido: %v", tt.expected, result)
			}
		})
	}
}
