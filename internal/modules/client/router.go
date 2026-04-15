package client

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute khusus admin/pengelolaan klien
func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	adminRoutes := routerGroup.Group("/admin")
	{
		// Akses ini nantinya bisa dilindungi middleware khusus SuperAdmin
		adminRoutes.POST("/clients", h.CreateClient)
	}
}
