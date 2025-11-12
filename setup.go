package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ubaidillahfaris/whatsapp.git/db"
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

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * 60 * 60

	r.Use(cors.New(config))
	r.OPTIONS("/*path", func(c *gin.Context) { c.Status(200) })

	routes.RegisterRoutes(r, mongo, manager)

	return r, nil
}
