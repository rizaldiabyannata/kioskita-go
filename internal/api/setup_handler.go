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
)

const configFilePath = "store_config.json"

type SetupHandler struct {
	adminStore        *store.AdminStore
	setupToken        string
	businessTemplates map[string]core.BusinessTemplate
}

func NewSetupHandler(adminStore *store.AdminStore, token string, templates map[string]core.BusinessTemplate) *SetupHandler {
	return &SetupHandler{
		adminStore:        adminStore,
		setupToken:        token,
		businessTemplates: templates,
	}
}

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

func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	providedToken := c.GetHeader("X-Setup-Token")
	if providedToken == "" || providedToken != h.setupToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token setup is not valid or missing."})
		return
	}
	count, err := h.adminStore.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing users."})
		return
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin user already exists. Please restart the application."})
		return
	}
	var payload core.AdminSetupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin := &core.Admin{
		ID:    uuid.New(),
		Name:  "Default Admin", // Or get from payload
		Email: payload.AdminEmail,
		Role:  core.SuperAdminRole, // First user is Super Admin
	}

	if err := admin.HashPassword(payload.AdminPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.adminStore.CreateAdmin(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user."})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Admin successfully created. Ready to configure the store."})
}

func (h *SetupHandler) ConfigureStore(c *gin.Context) {
	if _, err := os.Stat(configFilePath); err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "configuration already exists. Please restart the application."})
		return
	}
	var payload core.StoreSetupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, ok := h.businessTemplates[payload.BusinessTypeID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type of business not valid."})
		return
	}
	config := core.StoreConfig{
		StoreName:      payload.StoreName,
		BusinessType:   template.DisplayName,
		BusinessTypeID: payload.BusinessTypeID,
		SetupComplete:  true,
		ProductSchema:  template.Schema,
	}
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal configuration data."})
		return
	}
	if err := ioutil.WriteFile(configFilePath, configData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write configuration file."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Configuration successful. Please restart the application."})
}
