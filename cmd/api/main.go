package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rizaldiabyannata/kioskita-go/internal/api"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/service"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func loadBusinessTemplates() (map[string]core.BusinessTemplate, error) {
	templates := make(map[string]core.BusinessTemplate)
	templateDir := "./templates"
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca direktori template: %w", err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(templateDir, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Printf("Peringatan: Gagal membaca file template %s: %v", file.Name(), err)
				continue
			}
			var template core.BusinessTemplate
			if err := json.Unmarshal(data, &template); err != nil {
				log.Printf("Peringatan: Gagal parse file template %s: %v", file.Name(), err)
				continue
			}
			templateID := strings.TrimSuffix(file.Name(), ".json")
			templates[templateID] = template
			log.Printf("Template dimuat: '%s' (%s)", template.DisplayName, templateID)
		}
	}
	return templates, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan, menggunakan environment variables sistem.")
	}

	businessTemplates, err := loadBusinessTemplates()
	if err != nil {
		log.Fatalf("Kritis: Gagal memuat template bisnis: %v", err)
	}

	var db *gorm.DB
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "postgres" // Default to postgres
	}

	log.Printf("Menggunakan database driver: %s", dbDriver)

	switch dbDriver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSL_MODE"),
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open("kioskita_dev.db"), &gorm.Config{})
	default:
		log.Fatalf("Database driver tidak didukung: %s", dbDriver)
	}

	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	// Basic connectivity message
	log.Println("Berhasil terhubung ke database!")

	// Run database migrations
	if err := store.AutoMigrateDB(db); err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}
	log.Println("Migrasi database berhasil.")

	router := gin.Default()
	adminStore := store.NewAdminStore(db)

	_, configErr := os.Stat("store_config.json")
	adminCount, dbErr := adminStore.Count()
	if dbErr != nil {
		log.Fatalf("Kritis: Gagal memeriksa database saat startup: %v", dbErr)
	}

	v1 := router.Group("/api/v1")

	if adminCount == 0 {

		setupToken := uuid.NewString()
		log.Println("===================================================================")
		log.Println("===== APLIKASI DALAM MODE SETUP ADMIN =====")
		log.Printf("===> SETUP TOKEN: %s", setupToken)
		log.Println("Gunakan token di atas pada header 'X-Setup-Token' untuk otorisasi.")
		log.Println("===================================================================")
		setupHandler := api.NewSetupHandler(adminStore, setupToken, businessTemplates)
		v1.POST("/setup/admin", setupHandler.CreateAdmin)
		v1.GET("/setup/business-types", setupHandler.GetBusinessTypes)

	} else {
		// Admin sudah ada. Tentukan mode berdasarkan keberadaan dan kelengkapan store_config.json
		if configErr != nil {
			// File konfigurasi tidak ada => Mode Konfigurasi
			log.Println("===================================================================")
			log.Println("===== APLIKASI DALAM MODE KONFIGURASI TOKO =====")
			log.Println("Admin sudah ada, menunggu konfigurasi toko.")
			log.Println("===================================================================")
			setupHandler := api.NewSetupHandler(adminStore, "", businessTemplates)
			adminHandler := api.NewAdminHandler(adminStore)
			adminHandler.RegisterRoutes(v1)
			v1.GET("/setup/business-types", setupHandler.GetBusinessTypes)
			configRoute := v1.Group("/setup/config")
			configRoute.Use(api.AuthMiddleware())
			{
				configRoute.POST("/", setupHandler.ConfigureStore)
			}
		} else {
			// File ada: cek kelengkapan isian
			var appConfig core.StoreConfig
			configData, err := ioutil.ReadFile("store_config.json")
			if err != nil {
				log.Fatalf("Gagal membaca file konfigurasi: %v", err)
			}
			if err := json.Unmarshal(configData, &appConfig); err != nil {
				log.Fatalf("File konfigurasi rusak: %v", err)
			}

			if !appConfig.SetupComplete || appConfig.BusinessTypeID == "" {
				// Konfigurasi belum lengkap => Mode Konfigurasi
				log.Println("===================================================================")
				log.Println("===== APLIKASI DALAM MODE KONFIGURASI TOKO =====")
				log.Println("Admin sudah ada, menunggu konfigurasi toko (config belum lengkap).")
				log.Println("===================================================================")
				setupHandler := api.NewSetupHandler(adminStore, "", businessTemplates)
				adminHandler := api.NewAdminHandler(adminStore)
				adminHandler.RegisterRoutes(v1)
				v1.GET("/setup/business-types", setupHandler.GetBusinessTypes)
				configRoute := v1.Group("/setup/config")
				configRoute.Use(api.AuthMiddleware())
				{
					configRoute.POST("/", setupHandler.ConfigureStore)
				}
			} else {
				// Konfigurasi lengkap => Mode Operasional
				log.Println("Aplikasi berjalan dalam MODE OPERASIONAL.")

				// Instantiate Stores
				productStore := store.NewProductStore(db)
				orderStore := store.NewOrderStore(db)
				categoryStore := store.NewCategoryStore(db)
				voucherStore := store.NewVoucherStore(db)

				// Instantiate Services
				orderService := service.NewOrderService(db, orderStore, productStore, voucherStore)

				// Instantiate Handlers
				adminHandler := api.NewAdminHandler(adminStore)
				productHandler := api.NewProductHandler(productStore, categoryStore)
				orderHandler := api.NewOrderHandler(orderService, orderStore)

				// Register Routes
				adminHandler.RegisterRoutes(v1)
				productHandler.RegisterRoutes(v1)
				orderHandler.RegisterRoutes(v1)
			}
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	appPort := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	log.Printf("Server berjalan di port %s", appPort)
	if err := router.Run(appPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
