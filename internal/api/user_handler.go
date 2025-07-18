package api

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rizaldiabyannata/kioskita-go/internal/auth"
	"github.com/rizaldiabyannata/kioskita-go/internal/core"
	"github.com/rizaldiabyannata/kioskita-go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	store *store.UserStore
}

func NewUserHandler(store *store.UserStore) *UserHandler {
	return &UserHandler{store: store}
}

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
	router.POST("/logout", h.Logout)
}

func (h *UserHandler) Login(c *gin.Context) {
	var payload core.LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.store.GetByEmail(payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan server"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}
	token, err := auth.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.SetCookie("token", token, 86400, "/", os.Getenv("APP_DOMAIN"), true, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login berhasil"})
}

func (h *UserHandler) Logout(c *gin.Context) {

	c.SetCookie("token", "", -1, "/", os.Getenv("APP_DOMAIN"), true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout berhasil"})
}
