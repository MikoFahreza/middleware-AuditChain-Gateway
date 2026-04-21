package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	authRoutes := routerGroup.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
	}
}
