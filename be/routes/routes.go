package routes

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/handlers"
	"github.com/ubaidillahfaris/whatsapp.git/internal/app"
	"github.com/ubaidillahfaris/whatsapp.git/middlewares"
	"github.com/ubaidillahfaris/whatsapp.git/services"
)

func RegisterRoutes(r *gin.Engine, mongo *db.MongoService, manager *services.WhatsAppManager, container interface{}) {

	// Health check endpoint (for Docker health checks)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Authentication routes
	authHandler := handlers.NewAuthenticateHandler()
	auth := r.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			authHandler.Register(mongo, c)
		})
		auth.POST("/login", func(c *gin.Context) {
			authHandler.Authenticate(c)
		})

		auth.GET("/check", func(c *gin.Context) {
			authHandler.CheckAuth(c)
		})
	}

	sync := r.Group("/sync")
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			sync.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
		} else {
			sync.Use(middlewares.JWTAuthMiddleware())
		}
	} else {
		sync.Use(middlewares.JWTAuthMiddleware())
	}
	{
		sync.POST("/app", handlers.NewSyncHandler().SyncApp)
	}

	r.POST("/ping", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Baca public key
		pubKeyBytes, err := os.ReadFile("keys/go-wakey_public.pem")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
		if err != nil {
			c.JSON(500, gin.H{"error": "Invalid public key"})
			return
		}

		// Verifikasi token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return pubKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		username := claims["username"]

		c.JSON(200, gin.H{"message": "ok", "username": username})
	})

	// Device routes
	deviceHandler := handlers.NewDeviceHandler(mongo)

	device := r.Group("/devices")
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			device.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
		} else {
			device.Use(middlewares.JWTAuthMiddleware())
		}
	} else {
		device.Use(middlewares.JWTAuthMiddleware())
	}
	{
		device.POST("", deviceHandler.CreateDevice)
		device.GET("", deviceHandler.ListDevices)
		device.GET(":id", deviceHandler.GetDevice)
		device.PUT(":id", deviceHandler.UpdateDevice)
		device.DELETE(":id", deviceHandler.DeleteDevice)
	}

	// WhatsApp routes
	whatsapp := handlers.NewWhatsAppHandler()
	wa := r.Group("/whatsapp")
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			wa.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
		} else {
			wa.Use(middlewares.JWTAuthMiddleware())
		}
	} else {
		wa.Use(middlewares.JWTAuthMiddleware())
	}
	{
		wa.GET("/:device/qrcode", whatsapp.GenerateQR)
		wa.GET("/:device/status", whatsapp.GetStatus)
		wa.GET("/:device/disconnect", whatsapp.Disconnect)
		wa.GET("/:device/contacts", whatsapp.ListContacts)
		wa.GET("/:device/groups", whatsapp.ListGroups)
	}

	// Quick Response routes
	qrHandler := handlers.NewQuickResponseHandler()
	qr := r.Group("/quick_response")
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			qr.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
		} else {
			qr.Use(middlewares.JWTAuthMiddleware())
		}
	} else {
		qr.Use(middlewares.JWTAuthMiddleware())
	}
	{
		qr.GET("/", qrHandler.GetAll)
		qr.DELETE("/:id", qrHandler.DeleteId)
	}

	// Send Message routes
	msg := r.Group("/send_message")
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			msg.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
		} else {
			msg.Use(middlewares.JWTAuthMiddleware())
		}
	} else {
		msg.Use(middlewares.JWTAuthMiddleware())
	}
	{
		msg.POST("/:device", whatsapp.SendMessage)
	}

	// API Key routes (JWT protected for management)
	// Only register if container is provided (new architecture)
	if container != nil {
		if appContainer, ok := container.(*app.Container); ok {
			// Create API Key handler
			apiKeyHandler := handlers.NewAPIKeyHandler(
				appContainer.GenerateAPIKeyUC,
				appContainer.ListAPIKeysUC,
				appContainer.RevokeAPIKeyUC,
				appContainer.UpdateAPIKeyUC,
			)

			// API Key management endpoints (requires authentication via JWT or API Key)
			apiKeyGroup := r.Group("/api-keys")
			apiKeyGroup.Use(middlewares.APIKeyOrJWTMiddleware(appContainer.ValidateAPIKeyUC))
			{
				apiKeyGroup.POST("", apiKeyHandler.GenerateKey)       // Generate new API key
				apiKeyGroup.GET("", apiKeyHandler.ListKeys)           // List all user's API keys
				apiKeyGroup.GET("/:id", apiKeyHandler.GetKey)         // Get specific API key
				apiKeyGroup.PUT("/:id", apiKeyHandler.UpdateKey)      // Update API key
				apiKeyGroup.DELETE("/:id", apiKeyHandler.RevokeKey)   // Revoke (delete) API key
			}

			// API Key test endpoint (requires API Key authentication via X-API-Key header)
			apiKeyTestGroup := r.Group("/api-keys")
			{
				apiKeyTestGroup.POST("/test", middlewares.APIKeyMiddleware(appContainer.ValidateAPIKeyUC), apiKeyHandler.TestKey)
			}
		}
	}

}
