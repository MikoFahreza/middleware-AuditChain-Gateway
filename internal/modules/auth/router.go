package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes adalah pintu masuk API khusus untuk urusan Autentikasi
func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	// Modul ini membuat sub-grup "/auth" di bawah grup utama
	authRoutes := routerGroup.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
	}
}
