package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/internal/app"
	"github.com/ubaidillahfaris/whatsapp.git/routes"
	"github.com/ubaidillahfaris/whatsapp.git/services"
)

// setup mengembalikan *gin.Engine dan error
func setup() (*gin.Engine, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment")
	}

	mongo, err := db.InitMongoService()
	if err != nil {
		return nil, err
	}

	manager := services.GetWhatsAppManager()

	// Initialize container for Clean Architecture components (API Keys, etc.)
	container, err := app.NewContainer(context.Background())
	if err != nil {
		log.Printf("⚠️  Failed to initialize container: %v (API key routes will not be available)", err)
		container = nil
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma", "X-API-Key"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * 60 * 60

	r.Use(cors.New(config))
	r.OPTIONS("/*path", func(c *gin.Context) { c.Status(200) })

	// Pass container for API key routes
	routes.RegisterRoutes(r, mongo, manager, container)

	return r, nil
}
