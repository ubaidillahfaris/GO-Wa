package examples

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/internal/app"
	"github.com/ubaidillahfaris/whatsapp.git/routes"
)

// SetupWithContainer demonstrates how to set up the application using the Container
// This is an example of the new Clean Architecture approach
//
// To use this in main.go, replace the setup() function with:
//
//   func main() {
//       r, err := examples.SetupWithContainer(context.Background())
//       if err != nil {
//           log.Fatalf("❌ Setup failed: %v", err)
//       }
//
//       port := os.Getenv("PORT")
//       if port == "" {
//           port = "3000"
//       }
//
//       log.Printf("🚀 Server running on :%s", port)
//       if err := r.Run(":" + port); err != nil {
//           log.Fatalf("❌ Failed to run server: %v", err)
//       }
//   }
func SetupWithContainer(ctx context.Context) (*gin.Engine, error) {
	// Initialize container with all dependencies
	container, err := app.NewContainer(ctx)
	if err != nil {
		return nil, err
	}

	// Create Gin router
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma", "X-API-Key"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * 60 * 60

	r.Use(cors.New(config))
	r.OPTIONS("/*path", func(c *gin.Context) { c.Status(200) })

	// Register routes with container
	// Note: This passes nil for mongo and manager for backward compatibility
	// These can be removed once all routes are migrated to use the container
	routes.RegisterRoutes(r, nil, nil, container)

	return r, nil
}
