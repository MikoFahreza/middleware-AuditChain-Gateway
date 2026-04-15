package api

import (
	"go-blockchain-api/internal/modules/audit"
	"go-blockchain-api/internal/modules/auth"
	"go-blockchain-api/internal/modules/client" // 👈 IMPORT BARU
	"go-blockchain-api/internal/modules/ingestion"

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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := router.Group("/api")

	auth.RegisterRoutes(apiGroup, authHandler)
	client.RegisterRoutes(apiGroup, clientHandler)
	audit.RegisterRoutes(apiGroup, auditHandler)

	apiV1 := apiGroup.Group("/v1")
	ingestion.RegisterRoutes(apiV1, ingestionHandler)

	return router
}
