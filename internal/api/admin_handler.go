package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rizaldiabyannata/kioskita-go/internal/auth"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"gorm.io/gorm"
)

// AdminHandler handles authentication for admin users.
type AdminHandler struct {
	store *store.AdminStore
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(store *store.AdminStore) *AdminHandler {
	return &AdminHandler{store: store}
}

// RegisterRoutes registers the authentication routes.
func (h *AdminHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
	router.POST("/logout", h.Logout)
}

// Login handles the admin login process.
func (h *AdminHandler) Login(c *gin.Context) {
	var payload core.LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.store.FindAdminByEmail(payload.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan server"})
		return
	}

	if !admin.CheckPassword(payload.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	token, err := auth.GenerateToken(admin.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	// Use environment variables for cookie settings
	domain := os.Getenv("APP_DOMAIN")
	secure := os.Getenv("GIN_MODE") != "debug" // Secure cookies in production

	c.SetCookie("token", token, 86400, "/", domain, secure, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login berhasil", "token": token})
}

// Logout handles the admin logout process.
func (h *AdminHandler) Logout(c *gin.Context) {
	domain := os.Getenv("APP_DOMAIN")
	secure := os.Getenv("GIN_MODE") != "debug"

	c.SetCookie("token", "", -1, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout berhasil"})
}
