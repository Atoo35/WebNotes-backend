package configurations

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/lpernett/godotenv"
)

func LoadConfig() error {
	// Load configuration from file json
	// err := loadPermissions()
	// if err != nil {
	// 	return err
	// }

	dir, err := os.Getwd()
	if err != nil {
		return errors.New("error getting current directory")
	}
	// load env configs using go-dotenv from .env file
	filePath := filepath.Join(dir, ".env")
	err = godotenv.Load(filePath)
	if err != nil {
		log.Fatal("Error loading.env file")
	}
	return nil
}
