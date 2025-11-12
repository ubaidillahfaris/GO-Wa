package routes

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/handlers"
	"github.com/ubaidillahfaris/whatsapp.git/middlewares"
	"github.com/ubaidillahfaris/whatsapp.git/services"
)

func RegisterRoutes(r *gin.Engine, mongo *db.MongoService, manager *services.WhatsAppManager) {

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
	sync.Use(middlewares.JWTAuthMiddleware())
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
	device.Use(middlewares.JWTAuthMiddleware())
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
	{
		wa.GET("/:device/qrcode", whatsapp.GenerateQR)
		wa.GET("/:device/disconnect", whatsapp.Disconnect)
		wa.GET("/:device/contacts", whatsapp.ListContacts)
		wa.GET("/:device/groups", whatsapp.ListGroups)
	}

	// Quick Response routes
	qrHandler := handlers.NewQuickResponseHandler()
	qr := r.Group("/quick_response")
	{
		qr.GET("/", qrHandler.GetAll)
		qr.DELETE("/:id", qrHandler.DeleteId)
	}

	// Send Message routes

	msg := r.Group("/send_message")
	{
		msg.POST("/:device", whatsapp.SendMessage)
	}

}
