package models

import (
	"database/sql"
	"errors"
	"time"
)

// ErrSKUNoData es un error que indica que el SKU no tiene datos disponibles
// Este error no debe causar reintentos en Pub/Sub
var ErrSKUNoData = errors.New("SKU sin datos disponibles")

// IsNoDataError verifica si un error es de tipo "sin datos"
func IsNoDataError(err error) bool {
	return errors.Is(err, ErrSKUNoData)
}

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
	IDRemision     int64  `json:"idRemision"`
	IDItemRemision int64  `json:"idItemRemision"`
	OrderNumber    string `json:"orderNumber"`
	SKU            string `json:"sku"`
}

// ItemData representa los datos del item en Firestore (colección items-data-collection)
type ItemData struct {
	SKU            string `firestore:"sku" json:"sku"`
	Color          string `firestore:"color" json:"color"`
	EAN            string `firestore:"ean" json:"ean"`
	ImageURL       string `firestore:"imageURL" json:"imageURL"`
	ItemGroup      string `firestore:"itemGroup" json:"itemGroup"`
	ParentSKU      int64  `firestore:"parentSKU" json:"parentSKU"`
	ProductName    string `firestore:"productName" json:"productName"`
	Seccion        string `firestore:"seccion" json:"seccion"`
	TamanoUnico    string `firestore:"tamanoUnico" json:"tamanoUnico"`
	TextoAdicional string `firestore:"textoAdicional" json:"textoAdicional"`
	// HasData indica si este SKU tiene datos válidos (true) o si está marcado como "no encontrado" (false)
	HasData bool `firestore:"hasData" json:"hasData"`
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

// ItemRemisionDB representa el modelo real de la tabla ItemRemision en MySQL
type ItemRemisionDB struct {
	IDItemRemision              int64           `db:"IDItemRemision" json:"id_item_remision"`
	IDRemision                  sql.NullInt64   `db:"IDRemision" json:"id_remision,omitempty"`
	EAN1                        sql.NullString  `db:"EAN1" json:"ean1,omitempty"`
	EAN2                        sql.NullString  `db:"EAN2" json:"ean2,omitempty"`
	EAN3                        sql.NullString  `db:"EAN3" json:"ean3,omitempty"`
	EAN4                        sql.NullString  `db:"EAN4" json:"ean4,omitempty"`
	EAN5                        sql.NullString  `db:"EAN5" json:"ean5,omitempty"`
	FkDepartment                sql.NullString  `db:"fk_department" json:"fk_department,omitempty"`
	TiendaSurtido               string          `db:"tienda_surtido" json:"tienda_surtido"`
	FkStatus                    sql.NullString  `db:"fk_status" json:"fk_status,omitempty"`
	ImageURL                    sql.NullString  `db:"image_url" json:"image_url,omitempty"`
	MaterialName                string          `db:"material_name" json:"material_name"`
	MaterialSKU                 string          `db:"material_sku" json:"material_sku"`
	OrderQuantity               int64           `db:"order_quantity" json:"order_quantity"`
	StockAvailability           sql.NullInt64   `db:"stock_availability" json:"stock_availability,omitempty"`
	TotalPrice                  float64         `db:"total_price" json:"total_price"`
	IDItemRemisionRelacionada   sql.NullInt64   `db:"IDItemRemisionRelacionada" json:"id_item_remision_relacionada,omitempty"`
	CreatedAt                   sql.NullString  `db:"CreatedAt" json:"created_at,omitempty"`
	IDProcesoLogistico          sql.NullInt64   `db:"IDProcesoLogistico" json:"id_proceso_logistico,omitempty"`
	IDTipoRelaciona             sql.NullInt64   `db:"IDTipoRelaciona" json:"id_tipo_relaciona,omitempty"`
	IDMotivosRechazo            sql.NullInt64   `db:"IDMotivosRechazo" json:"id_motivos_rechazo,omitempty"`
	IndicadorRecepcionLogistica sql.NullBool    `db:"IndicadorRecepcionLogistica" json:"indicador_recepcion_logistica,omitempty"`
	IDSistemaOrigem             sql.NullInt64   `db:"IDSistemaOrigem" json:"id_sistema_origem,omitempty"`
	NoGuia                      sql.NullString  `db:"NoGuia" json:"no_guia,omitempty"`
	Seller                      sql.NullString  `db:"Seller" json:"seller,omitempty"`
	Activo                      sql.NullBool    `db:"Activo" json:"activo,omitempty"`
	TipoFelicitacion            sql.NullString  `db:"tipo_felicitacion" json:"tipo_felicitacion,omitempty"`
	MensajeFelicitacion         sql.NullString  `db:"mensaje_felicitacion" json:"mensaje_felicitacion,omitempty"`
	EntregadoSeccion            sql.NullString  `db:"entregado_seccion" json:"entregado_seccion,omitempty"`
	GrupoConsolidacion          sql.NullString  `db:"grupo_consolidacion" json:"grupo_consolidacion,omitempty"`
	AltoValor                   sql.NullString  `db:"AltoValor" json:"alto_valor,omitempty"`
	Estanteria                  sql.NullString  `db:"Estanteria" json:"estanteria,omitempty"`
	TiendaDestino               sql.NullString  `db:"tienda_destino" json:"tienda_destino,omitempty"`
	Carrier                     sql.NullString  `db:"Carrier" json:"carrier,omitempty"`
	ErrorSOMS                   sql.NullString  `db:"Error_SOMS" json:"error_soms,omitempty"`
	IdentificadorSOMS           sql.NullString  `db:"Identificador_SOMS" json:"identificador_soms,omitempty"`
	ConfSOMS                    sql.NullString  `db:"Conf_SOMS" json:"conf_soms,omitempty"`
	CodigoOrigenWMS             sql.NullString  `db:"codigo_origen_WMS" json:"codigo_origen_wms,omitempty"`
	FechaManifestoWMS           sql.NullString  `db:"fecha_manifesto_WMS" json:"fecha_manifesto_wms,omitempty"`
	IDTipoEntrega               sql.NullInt64   `db:"IDTipoEntrega" json:"id_tipo_entrega,omitempty"`
	IDRemisionTrackingStep      sql.NullInt64   `db:"IDRemisionTrackingStep" json:"id_remision_tracking_step,omitempty"`
	ClickCollectTienda          sql.NullString  `db:"Click_Collect_Tienda" json:"click_collect_tienda,omitempty"`
	OrderNumber                 sql.NullString  `db:"order_number" json:"order_number,omitempty"`
	Expreso                     sql.NullBool    `db:"Expreso" json:"expreso,omitempty"`
	FechaAsignacionEmpleado     sql.NullString  `db:"FechaAsignacionEmpleado" json:"fecha_asignacion_empleado,omitempty"`
	IDEmpleado                  sql.NullInt64   `db:"IDEmpleado" json:"id_empleado,omitempty"`
	TiempoSurtido               sql.NullInt64   `db:"TiempoSurtido" json:"tiempo_surtido,omitempty"`
	OrfanDepartamento           sql.NullBool    `db:"OrfanDepartamento" json:"orfan_departamento"`
	FechaSurtido                sql.NullString  `db:"FechaSurtido" json:"fecha_surtido,omitempty"`
	FinalizadoAppSurtido        sql.NullBool    `db:"FinalizadoAppSurtido" json:"finalizado_app_surtido"`
	FechaFinalizado             sql.NullString  `db:"FechaFinalizado" json:"fecha_finalizado,omitempty"`
	FechaAsignacionTienda       sql.NullString  `db:"FechaAsignacionTienda" json:"fecha_asignacion_tienda,omitempty"`
	Folio                       sql.NullInt64   `db:"Folio" json:"folio,omitempty"`
	FechaEstatus                sql.NullString  `db:"FechaEstatus" json:"fecha_estatus,omitempty"`
	TiempoSurtidoTienda         sql.NullInt64   `db:"TiempoSurtidoTienda" json:"tiempo_surtido_tienda,omitempty"`
	TiempoXD                    sql.NullInt64   `db:"TiempoXD" json:"tiempo_xd,omitempty"`
	IDTienda                    sql.NullInt64   `db:"IDTienda" json:"id_tienda,omitempty"`
	SurtidoCedis                sql.NullBool    `db:"SurtidoCedis" json:"surtido_cedis,omitempty"`
	IDDeleted                   sql.NullInt64   `db:"IDDeleted" json:"id_deleted,omitempty"`
	StatusGestor                sql.NullInt64   `db:"StatusGestor" json:"status_gestor,omitempty"`
	TrazaLOGS                   sql.NullString  `db:"TrazaLOGS" json:"traza_logs,omitempty"`
	PackedUser                  sql.NullInt64   `db:"PackedUser" json:"packed_user,omitempty"`
	IVA                         sql.NullFloat64 `db:"IVA" json:"iva,omitempty"`
	SuggestedPackageName        sql.NullString  `db:"suggested_package_name" json:"suggested_package_name,omitempty"`
	SuggestedPackageSKU         sql.NullString  `db:"suggested_package_sku" json:"suggested_package_sku,omitempty"`
	TotalQuantity               sql.NullInt64   `db:"TotalQuantity" json:"total_quantity,omitempty"`
	IsReship                    sql.NullBool    `db:"IsReship" json:"is_reship,omitempty"`
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
