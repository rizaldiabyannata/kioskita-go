package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const configFilePath = "store_config.json"

type SetupHandler struct {
	userStore         *store.UserStore
	setupToken        string
	businessTemplates map[string]core.BusinessTemplate
}

func NewSetupHandler(userStore *store.UserStore, token string, templates map[string]core.BusinessTemplate) *SetupHandler {
	return &SetupHandler{
		userStore:         userStore,
		setupToken:        token,
		businessTemplates: templates,
	}
}

// GetBusinessTypes menyediakan daftar tipe bisnis yang tersedia untuk frontend.
func (h *SetupHandler) GetBusinessTypes(c *gin.Context) {
	type TemplateInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var availableTypes []TemplateInfo
	for id, template := range h.businessTemplates {
		availableTypes = append(availableTypes, TemplateInfo{ID: id, Name: template.DisplayName})
	}
	c.JSON(http.StatusOK, availableTypes)
}

// CreateAdmin menangani pembuatan pengguna admin pertama.
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	providedToken := c.GetHeader("X-Setup-Token")
	if providedToken == "" || providedToken != h.setupToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token setup tidak valid atau tidak ada."})
		return
	}
	count, err := h.userStore.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa database."})
		return
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin sudah pernah dibuat."})
		return
	}
	var payload core.AdminSetupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(payload.AdminPassword), bcrypt.DefaultCost)
	adminUser := &core.User{
		ID:           uuid.NewString(),
		Email:        payload.AdminEmail,
		PasswordHash: string(hashedPassword),
	}
	if err := h.userStore.Create(adminUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pengguna admin."})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Admin berhasil dibuat. Silakan restart aplikasi, lalu login untuk melanjutkan setup toko."})
}

// ConfigureStore menangani penyimpanan detail konfigurasi toko.
func (h *SetupHandler) ConfigureStore(c *gin.Context) {
	if _, err := os.Stat(configFilePath); err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Konfigurasi toko sudah pernah dibuat."})
		return
	}
	var payload core.StoreSetupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, ok := h.businessTemplates[payload.BusinessTypeID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe bisnis tidak valid."})
		return
	}
	config := core.StoreConfig{
		StoreName:     payload.StoreName,
		BusinessType:  template.DisplayName,
		SetupComplete: true,
		ProductSchema: template.Schema,
	}
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat file konfigurasi."})
		return
	}
	if err := ioutil.WriteFile(configFilePath, configData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file konfigurasi."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Konfigurasi toko berhasil disimpan. Silakan restart aplikasi untuk masuk ke mode operasional penuh."})
}
