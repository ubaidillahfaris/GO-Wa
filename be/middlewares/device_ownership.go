package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"go.mongodb.org/mongo-driver/bson"
)

// DeviceOwnershipMiddleware checks if the user owns the device specified in the route parameter
func DeviceOwnershipMiddleware(mongo *db.MongoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get username from JWT (set by JWTAuthMiddleware)
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Get device name from route parameter
		deviceName := c.Param("device")
		if deviceName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "device name is required"})
			c.Abort()
			return
		}

		// Check if device exists and belongs to the user
		filter := bson.M{
			"name":  deviceName,
			"owner": username.(string),
		}

		var skip int64 = 0
		var limit int64 = 1
		devices, err := mongo.FindAll(c.Request.Context(), "devices", filter, &skip, &limit)
		if err != nil || len(devices) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "device not found or access denied"})
			c.Abort()
			return
		}

		// Store device info in context for handler use
		c.Set("device", devices[0])
		c.Next()
	}
}
