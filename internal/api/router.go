package api

import (
	"go-blockchain-api/internal/modules/audit"
	"go-blockchain-api/internal/modules/auth"
	"go-blockchain-api/internal/modules/client"
	"go-blockchain-api/internal/modules/ingestion"

	"github.com/gin-contrib/cors"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-blockchain-api/docs"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	ingestionHandler *ingestion.Handler,
	auditHandler *audit.Handler,
	authHandler *auth.Handler,
	clientHandler *client.Handler,
) *gin.Engine {

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true // Mengizinkan request dari semua origin (termasuk localhost:3000)
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "api-key"}
	router.Use(cors.New(corsConfig))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := router.Group("/api")

	auth.RegisterRoutes(apiGroup, authHandler)
	client.RegisterRoutes(apiGroup, clientHandler)
	audit.RegisterRoutes(apiGroup, auditHandler)

	apiV1 := apiGroup.Group("/v1")
	ingestion.RegisterRoutes(apiV1, ingestionHandler)

	return router
}
