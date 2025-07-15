package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rizaldiabyannata/kioskita-go/internal/api"   // Ganti 'kioskita' dengan nama modul Go Anda jika berbeda
	"github.com/rizaldiabyannata/kioskita-go/internal/store" // Ganti 'kioskita' dengan nama modul Go Anda jika berbeda

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer db.Close()
	log.Println("Berhasil terhubung ke database!")

	router := gin.Default()

	// Setup lapisan store dan handler
	productStore := store.NewProductStore(db)
	productHandler := api.NewProductHandler(productStore)

	// Buat grup rute untuk API versi 1
	v1 := router.Group("/api/v1")
	{
		// Daftarkan semua rute produk di dalam grup v1
		productHandler.RegisterRoutes(v1)
	}

	// Endpoint untuk mengecek status server
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	appPort := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	log.Printf("Server berjalan di port %s", appPort)
	if err := router.Run(appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
