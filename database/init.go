package database

import (
    "os"
	"fmt"
	"golang-app/app/models"

    _ "github.com/jinzhu/gorm/dialects/postgres"
    "github.com/jinzhu/gorm"

)

var DB *gorm.DB

func Init() {
	dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USER")
    dbName := os.Getenv("DB_NAME")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbSSLMode := os.Getenv("DB_SSL_MODE")

    dbConnection := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
        dbHost,
        dbPort,
        dbUser,
        dbName,
        dbPassword,
        dbSSLMode,
    )

    var errDb error
    // Connect to PostgreSQL database
    DB, errDb = gorm.Open("postgres", dbConnection)
    if errDb != nil {
        panic("failed to connect database")
    }

	DB.AutoMigrate(&models.User{})
}