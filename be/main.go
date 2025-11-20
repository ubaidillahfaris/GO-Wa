package main

import (
	"log"
	"os"
)

func main() {
	r, err := setup()
	if err != nil {
		log.Fatalf("❌ Setup failed: %v", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("❌ JWT_SECRET not set in .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to run server: %v", err)
	}
}
