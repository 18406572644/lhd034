package main

import (
	"cartridge-archive/database"
	"cartridge-archive/routes"
	"log"
	"os"
)

func main() {
	log.Println("Starting Cartridge Archive Server...")

	database.InitDB()
	log.Println("Database initialized")

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	log.Println("API Documentation:")
	log.Println("  GET    /api/health")
	log.Println("  GET    /api/cartridges")
	log.Println("  POST   /api/cartridges")
	log.Println("  GET    /api/cartridges/:id")
	log.Println("  PUT    /api/cartridges/:id")
	log.Println("  DELETE /api/cartridges/:id")
	log.Println("  GET    /api/playthroughs")
	log.Println("  POST   /api/playthroughs")
	log.Println("  GET    /api/wishlist")
	log.Println("  GET    /api/borrows")
	log.Println("  GET    /api/statistics/overview")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
