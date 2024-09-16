package database

import (
	"fmt"
	"golang-app/app/models"
	"log"
	"os"

	// _ "github.com/jinzhu/gorm/dialects/postgres" // Postgres
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // MYSQL

)

var DB *gorm.DB

func Init() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	// dbSSLMode := os.Getenv("DB_SSL_MODE")

	dbConnection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	var errDb error
	// Connect to PostgreSQL database
	DB, errDb = gorm.Open("mysql", dbConnection)
	if errDb != nil {
		panic("failed to connect database")
	}

	err := DB.AutoMigrate(&models.User{}, &models.Mapping{}, &models.Kapal{}, &models.IPKapal{}, &models.VesselRecord{}).Error
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Set up foreign key constraint
	if err := DB.Model(&models.IPKapal{}).AddForeignKey("call_sign", "kapals(call_sign)", "CASCADE", "CASCADE").Error; err != nil {
		log.Fatal("Failed to set up foreign key:", err)
	}
	if err := DB.Model(&models.VesselRecord{}).AddForeignKey("call_sign", "kapals(call_sign)", "CASCADE", "CASCADE").Error; err != nil {
		log.Fatal("Failed to set up foreign key:", err)
	}

	DB.LogMode(true)
}
