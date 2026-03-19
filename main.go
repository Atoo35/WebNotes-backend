package main

import (
	"log"
	"os"
	"strings"
	"time"

	notion "github.com/Atoo35/WebNotes-backend/src/clients/notion"
	"github.com/Atoo35/WebNotes-backend/src/configurations"
	"github.com/Atoo35/WebNotes-backend/src/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/Atoo35/WebNotes-backend/src/utils/database"
	zlog "github.com/Atoo35/WebNotes-backend/src/utils/logger"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load configuration
	err := configurations.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// Initialize logger before anything else uses it
	zlog.InitLogger()

	database.SQLClient.NewDB()
	notion.NotionClient.NewNotionClient()

	// Create a Gin router instance
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:           strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:           []string{"*"},
		AllowCredentials:       false,
		AllowBrowserExtensions: true,
		MaxAge:                 12 * time.Hour,
	}))

	routes.SetupRoutes(r)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
