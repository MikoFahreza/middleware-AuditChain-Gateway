package api

import (
	"go-blockchain-api/internal/modules/audit"
	"go-blockchain-api/internal/modules/auth"
	"go-blockchain-api/internal/modules/ingestion"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	ingestionHandler *ingestion.Handler,
	auditHandler *audit.Handler,
	authHandler *auth.Handler,
) *gin.Engine {

	router := gin.Default()

	apiGroup := router.Group("/api")
	auth.RegisterRoutes(apiGroup, authHandler)
	apiV1 := apiGroup.Group("/v1")
	ingestion.RegisterRoutes(apiV1, ingestionHandler)
	audit.RegisterRoutes(apiGroup, auditHandler)

	return router
}
