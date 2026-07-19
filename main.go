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

	// Crear contexto para inicialización
	ctx := context.Background()

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

	// Inicializar Firestore store
	firestoreStore := store.NewFirestoreStore(firestoreClient, appCfg.Firestore.ItemsCollection)

	// Inicializar servicios
	itemService := services.NewItemService(firestoreStore, appCfg)

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
