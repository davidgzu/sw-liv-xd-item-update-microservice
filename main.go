package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"sw-liv-xd-item-update-microservice/internal/config"
	"sw-liv-xd-item-update-microservice/internal/handlers"
	"sw-liv-xd-item-update-microservice/internal/router"
	"sw-liv-xd-item-update-microservice/internal/services"
	"sw-liv-xd-item-update-microservice/internal/store"

	"cloud.google.com/go/firestore"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Cargar configuración
	appCfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error fatal al cargar la configuración: %v", err)
	}

	log.Printf("Aplicación iniciando en entorno: %s", appCfg.Env)
	log.Printf("Configuración de Firestore: ProjectID=%s, Database=%s",
		appCfg.Firestore.ProjectID, appCfg.Firestore.DatabaseID)
	log.Printf("Configuración de MySQL: %s", appCfg.Database.SafeSummary())

	// Crear contexto para inicialización
	ctx := context.Background()

	// Inicializar conexión MySQL
	dsn := appCfg.Database.DSN()
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Error fatal al conectar a MySQL: %v", err)
	}
	defer db.Close()

	// Configurar pool de conexiones
	db.SetMaxOpenConns(appCfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(appCfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(appCfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(appCfg.Database.ConnMaxIdleTime)

	log.Println("✅ Conexión a MySQL establecida correctamente")

	// Inicializar cliente de Firestore
	firestoreClient, err := firestore.NewClientWithDatabase(
		ctx,
		appCfg.Firestore.ProjectID,
		appCfg.Firestore.DatabaseID,
	)
	if err != nil {
		log.Fatalf("Error fatal al conectar a Firestore: %v", err)
	}
	defer firestoreClient.Close()

	log.Println("✅ Conexión a Firestore establecida correctamente")
	log.Printf("   Base de datos: %s", appCfg.Firestore.DatabaseID)
	log.Printf("   Colección items: %s", appCfg.Firestore.ItemsCollection)

	// Inicializar stores
	firestoreStore := store.NewFirestoreStore(firestoreClient, appCfg.Firestore.ItemsCollection)
	mysqlStore := store.NewItemStore(db)

	// Inicializar cliente de servicio externo
	externalServiceClient := store.NewExternalServiceClient(appCfg.External.ServiceURL)
	log.Printf("✅ Cliente de servicio externo configurado: %s", appCfg.External.ServiceURL)

	// Inicializar servicios
	itemService := services.NewItemService(firestoreStore, mysqlStore, externalServiceClient, appCfg)

	// Inicializar handlers
	itemHandler := handlers.NewItemHandler(itemService)

	// Configurar Echo
	e := echo.New()
	e.Server.ReadTimeout = appCfg.Server.ReadTimeout
	e.Server.WriteTimeout = appCfg.Server.WriteTimeout

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Configurar rutas
	router.SetupRoutes(e, itemHandler)

	// Iniciar servidor
	go func() {
		addr := fmt.Sprintf(":%s", appCfg.Server.Port)
		log.Printf("Servidor iniciando en el puerto %s", appCfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar servidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Deteniendo servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Servidor detenido correctamente")
}
