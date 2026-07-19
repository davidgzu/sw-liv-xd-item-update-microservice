package models

import "time"

// PubSubMessage representa el mensaje recibido de Google Cloud Pub/Sub
type PubSubMessage struct {
	Message struct {
		Data        string            `json:"data"`
		Attributes  map[string]string `json:"attributes"`
		MessageID   string            `json:"messageId"`
		PublishTime time.Time         `json:"publishTime"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// ItemUpdateRequest representa la solicitud de actualización de item desde Pub/Sub
type ItemUpdateRequest struct {
	SaveTable   string    `json:"saveTable"`
	SaveDataset string    `json:"saveDataset"`
	LogObject   LogObject `json:"logObject"`
}

// LogObject contiene la información del item a actualizar
type LogObject struct {
	IDRemision     int64  `json:"idRemision"`
	IDItemRemision int64  `json:"idItemRemision"`
	OrderNumber    string `json:"orderNumber"`
	SKU            string `json:"sku"`
}

// ItemRemision representa un item en la tabla item_remision
type ItemRemision struct {
	ID          int64     `db:"id" json:"id"`
	SKU         string    `db:"sku" json:"sku"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	Stock       int       `db:"stock" json:"stock"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ExternalServiceResponse representa la respuesta del servicio externo
type ExternalServiceResponse struct {
	SKU         string                 `json:"sku"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
	Data        map[string]interface{} `json:"data"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
}
