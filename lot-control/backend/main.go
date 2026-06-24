package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"lot-control/internal/db"
	"lot-control/internal/handler"
	"lot-control/internal/middleware"
)

func main() {
	database, err := db.Init("./data/lot_control.db")
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	jwtSecret := getEnv("JWT_SECRET", "lot-control-secret-2024")

	authH := handler.NewAuthHandler(database, jwtSecret)
	skuH := handler.NewSKUHandler(database)
	lotH := handler.NewLotHandler(database)
	userH := handler.NewUserHandler(database)
	jwtMW := middleware.JWT(jwtSecret)

	api := r.Group("/api")

	// Public
	api.POST("/auth/login", authH.Login)

	// Protected
	p := api.Group("/")
	p.Use(jwtMW)

	p.GET("/auth/me", authH.Me)

	p.GET("/skus", skuH.List)
	p.POST("/skus", skuH.Create)
	p.GET("/skus/:id", skuH.GetByID)
	p.PUT("/skus/:id", skuH.Update)
	p.DELETE("/skus/:id", skuH.Delete)

	p.GET("/lots", lotH.List)
	p.POST("/lots", lotH.Upsert)
	p.PUT("/lots/:id", lotH.Update)
	p.DELETE("/lots/:id", lotH.Delete)

	p.GET("/users", userH.List)
	p.POST("/users", userH.Create)
	p.DELETE("/users/:id", userH.Delete)

	port := getEnv("PORT", "8080")
	log.Printf("Server listening on :%s", port)
	log.Printf("Default login: admin / admin123")
	log.Fatal(r.Run(":" + port))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
