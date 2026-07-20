package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sw-liv-xd-item-update-microservice/internal/models"
)

func TestExternalServiceClient_GetItemDataBySKU(t *testing.T) {
	tests := []struct {
		name           string
		sku            string
		mockResponse   []models.ItemData
		mockStatusCode int
		expectError    bool
		expectedSKU    string
		description    string
	}{
		{
			name: "Respuesta exitosa con datos",
			sku:  "1033804373",
			mockResponse: []models.ItemData{
				{
					SKU:            "1033804373",
					ProductName:    "Apple iPhone 16 Plus, 256 Gb, Negro",
					Color:          "Negro",
					TamanoUnico:    "256 GB",
					EAN:            "2050073491340",
					ImageURL:       "https://example.com/image.jpg",
					ItemGroup:      "TELEFONIA LIBRE",
					ParentSKU:      1033804362,
					Seccion:        "632 - CELULARES",
					TextoAdicional: "IPHONE SOLI INSUMOS LIVERSTORE",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedSKU:    "1033804373",
			description:    "Debería retornar el primer item del array",
		},
		{
			name:           "Respuesta vacía",
			sku:            "999999",
			mockResponse:   []models.ItemData{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			description:    "Debería retornar nil cuando el array está vacío",
		},
		{
			name:           "Error 404",
			sku:            "999999",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
			description:    "Debería retornar error cuando el servicio retorna 404",
		},
		{
			name:           "Error 500",
			sku:            "123456",
			mockResponse:   nil,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			description:    "Debería retornar error cuando el servicio retorna 500",
		},
		{
			name: "SKU no viene en respuesta",
			sku:  "123456",
			mockResponse: []models.ItemData{
				{
					SKU:         "", // SKU vacío
					ProductName: "Producto sin SKU",
					Color:       "Azul",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedSKU:    "123456",
			description:    "Debería usar el SKU del parámetro si no viene en la respuesta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear servidor HTTP mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verificar que el SKU viene en el query param
				sku := r.URL.Query().Get("skus")
				if sku != tt.sku {
					t.Errorf("SKU en query param esperado: %s, obtenido: %s", tt.sku, sku)
				}

				// Configurar respuesta
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Crear cliente con la URL del servidor mock
			client := NewExternalServiceClient(server.URL)

			// Ejecutar el método
			ctx := context.Background()
			result, err := client.GetItemDataBySKU(ctx, tt.sku)

			// Verificar errores
			if tt.expectError && err == nil {
				t.Errorf("%s: se esperaba un error pero no ocurrió", tt.description)
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("%s: no se esperaba error pero ocurrió: %v", tt.description, err)
				return
			}

			// Verificar resultado
			if !tt.expectError && len(tt.mockResponse) > 0 {
				if result == nil {
					t.Errorf("%s: resultado no debería ser nil", tt.description)
					return
				}

				if result.SKU != tt.expectedSKU {
					t.Errorf("%s: SKU esperado: %s, obtenido: %s", tt.description, tt.expectedSKU, result.SKU)
				}

				if result.ProductName != tt.mockResponse[0].ProductName {
					t.Errorf("%s: ProductName esperado: %s, obtenido: %s", tt.description, tt.mockResponse[0].ProductName, result.ProductName)
				}
			}

			// Verificar que retorna nil cuando el array está vacío
			if !tt.expectError && len(tt.mockResponse) == 0 && result != nil {
				t.Errorf("%s: resultado debería ser nil cuando el array está vacío", tt.description)
			}
		})
	}
}

func TestExternalServiceClient_GetItemDataBySKU_EmptySKU(t *testing.T) {
	client := NewExternalServiceClient("http://example.com")
	ctx := context.Background()

	result, err := client.GetItemDataBySKU(ctx, "")

	if err == nil {
		t.Error("Se esperaba un error cuando el SKU está vacío")
	}

	if result != nil {
		t.Error("El resultado debería ser nil cuando el SKU está vacío")
	}
}
