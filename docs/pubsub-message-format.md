# Ejemplo de Mensaje Pub/Sub

Este archivo contiene el ejemplo del mensaje que llegará desde Google Cloud Pub/Sub.

## Formato del Mensaje

El mensaje de Pub/Sub llega en el siguiente formato:

```json
{
  "message": {
    "data": "base64_encoded_data",
    "messageId": "test-message-id-123",
    "publishTime": "2026-07-18T10:00:00Z",
    "attributes": {
      "source": "item-update-system"
    }
  },
  "subscription": "projects/your-project-id/subscriptions/item-update-subscription"
}
```

## Contenido Decodificado (campo `data`)

El campo `data` está codificado en base64. Al decodificarlo, contiene:

```json
{
  "saveTable": "UPDATE_ITEM_DESCRIPTION",
  "saveDataset": "mvp_analiticos",
  "logObject": {
    "idRemision": 115576358,
    "idItemRemision": 115576358,
    "orderNumber": "ORDER9-123003",
    "sku": "1000123456"
  }
}
```

## Campos del Mensaje

### ItemUpdateRequest
- `saveTable`: Tabla donde se guardará la información (ej: "UPDATE_ITEM_DESCRIPTION")
- `saveDataset`: Dataset de destino (ej: "mvp_analiticos")
- `logObject`: Objeto con la información del item

### LogObject
- `idRemision`: ID de la remisión
- `idItemRemision`: ID del item en la remisión
- `orderNumber`: Número de orden
- `sku`: SKU del producto a actualizar

## Probar Localmente

Puedes enviar el mensaje de ejemplo usando curl:

```bash
curl -X POST http://localhost:8080/api/v1/items/pubsub \
  -H "Content-Type: application/json" \
  -d @docs/pubsub-message-example.json
```

## Generar Base64 del Payload

Si necesitas generar un nuevo mensaje codificado en base64:

```bash
# Linux/macOS
echo -n '{"saveTable":"UPDATE_ITEM_DESCRIPTION","saveDataset":"mvp_analiticos","logObject":{"idRemision":115576358,"idItemRemision":115576358,"orderNumber":"ORDER9-123003","sku":"1000123456"}}' | base64

# Windows PowerShell
$text = '{"saveTable":"UPDATE_ITEM_DESCRIPTION","saveDataset":"mvp_analiticos","logObject":{"idRemision":115576358,"idItemRemision":115576358,"orderNumber":"ORDER9-123003","sku":"1000123456"}}'
[Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($text))
```

## Flujo de Procesamiento

1. Pub/Sub envía el mensaje al endpoint `/api/v1/items/pubsub`
2. El handler decodifica el campo `data` desde base64
3. Se parsea el JSON al struct `ItemUpdateRequest`
4. Se extrae el SKU desde `logObject.sku`
5. Se busca el item en la base de datos por SKU
6. Si existe, se actualiza; si no existe, se consulta el servicio externo
7. Se guarda/actualiza en la tabla `item_remision`
8. Se devuelve HTTP 200 para que Pub/Sub haga ACK del mensaje
