package store

import (
	"testing"

	"sw-liv-xd-item-update-microservice/internal/models"
)

func TestBuildMaterialName(t *testing.T) {
	store := &ItemStore{}

	tests := []struct {
		name        string
		itemData    *models.ItemData
		expected    string
		description string
	}{
		{
			name: "ProductName con Color y Talla",
			itemData: &models.ItemData{
				ProductName: "Apple iPhone 16 Plus, 256 Gb, Negro",
				Color:       "Negro",
				TamanoUnico: "256 GB",
			},
			expected:    "Apple iPhone 16 Plus, 256 Gb, Negro | Color: Negro | Talla: 256 GB",
			description: "Debe incluir ProductName + Color + Talla",
		},
		{
			name: "ProductName con Color sin Talla",
			itemData: &models.ItemData{
				ProductName: "Camisa de algodón",
				Color:       "Azul",
				TamanoUnico: "",
			},
			expected:    "Camisa de algodón | Color: Azul",
			description: "Solo incluye ProductName + Color",
		},
		{
			name: "ProductName con Talla sin Color",
			itemData: &models.ItemData{
				ProductName: "Pantalón casual",
				Color:       "",
				TamanoUnico: "32",
			},
			expected:    "Pantalón casual | Talla: 32",
			description: "Solo incluye ProductName + Talla",
		},
		{
			name: "Solo ProductName sin Color ni Talla",
			itemData: &models.ItemData{
				ProductName: "Mouse inalámbrico",
				Color:       "",
				TamanoUnico: "",
			},
			expected:    "Mouse inalámbrico",
			description: "Solo retorna ProductName",
		},
		{
			name: "Tenis deportivos completo",
			itemData: &models.ItemData{
				ProductName: "Tenis Nike Air Max",
				Color:       "Negro/Blanco",
				TamanoUnico: "27 MX",
			},
			expected:    "Tenis Nike Air Max | Color: Negro/Blanco | Talla: 27 MX",
			description: "Formato completo con todos los campos",
		},
		{
			name: "ProductName vacío",
			itemData: &models.ItemData{
				ProductName: "",
				Color:       "Rojo",
				TamanoUnico: "M",
			},
			expected:    "Color: Rojo | Talla: M",
			description: "Solo Color y Talla si ProductName está vacío",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.buildMaterialName(tt.itemData)
			if result != tt.expected {
				t.Errorf("\n%s\nEsperado: %s\nObtenido: %s", tt.description, tt.expected, result)
			}
		})
	}
}
