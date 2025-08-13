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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rizaldiabyannata/kioskita-go/internal/api"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
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
	userStore := store.NewUserStore(db)

	_, configErr := os.Stat("store_config.json")
	userCount, dbErr := userStore.Count()
	if dbErr != nil {
		log.Fatalf("Kritis: Gagal memeriksa database saat startup: %v", dbErr)
	}

	v1 := router.Group("/api/v1")

	if userCount == 0 {

		setupToken := uuid.NewString()
		log.Println("===================================================================")
		log.Println("===== APLIKASI DALAM MODE SETUP ADMIN =====")
		log.Printf("===> SETUP TOKEN: %s", setupToken)
		log.Println("Gunakan token di atas pada header 'X-Setup-Token' untuk otorisasi.")
		log.Println("===================================================================")
		setupHandler := api.NewSetupHandler(userStore, setupToken, businessTemplates)
		v1.POST("/setup/admin", setupHandler.CreateAdmin)
		v1.GET("/setup/business-types", setupHandler.GetBusinessTypes)

	} else if configErr != nil {

		log.Println("===================================================================")
		log.Println("===== APLIKASI DALAM MODE KONFIGURASI TOKO =====")
		log.Println("Admin sudah ada, menunggu konfigurasi toko.")
		log.Println("===================================================================")
		setupHandler := api.NewSetupHandler(userStore, "", businessTemplates)
		userHandler := api.NewUserHandler(userStore)
		v1.POST("/login", userHandler.Login)
		v1.GET("/setup/business-types", setupHandler.GetBusinessTypes)
		configRoute := v1.Group("/setup/config")
		configRoute.Use(api.AuthMiddleware())
		{
			configRoute.POST("/", setupHandler.ConfigureStore)
		}
	} else {

		log.Println("Aplikasi berjalan dalam MODE OPERASIONAL.")
		var appConfig core.StoreConfig
		configData, err := ioutil.ReadFile("store_config.json")
		if err != nil {
			log.Fatalf("Gagal membaca file konfigurasi: %v", err)
		}
		if err := json.Unmarshal(configData, &appConfig); err != nil {
			log.Fatalf("File konfigurasi rusak: %v", err)
		}

		// Inisialisasi semua store
		productStore := store.NewProductStore(db)
		mediaStore := store.NewMediaStore(db)
		// Berikan productStore ke NewOrderStore
		orderStore := store.NewOrderStore(db, productStore)

		// Inisialisasi semua handler
		productHandler := api.NewProductHandler(productStore, mediaStore, &appConfig)
		userHandler := api.NewUserHandler(userStore)
		// Berikan orderStore yang sudah lengkap ke NewOrderHandler
		orderHandler := api.NewOrderHandler(orderStore)

		// Daftarkan semua rute
		userHandler.RegisterRoutes(v1)
		productHandler.RegisterRoutes(v1)
		orderHandler.RegisterRoutes(v1)
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
