package ingestion

import "github.com/gin-gonic/gin"

func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	// Modul ini membuat rute untuk "/logs"
	logsRoutes := routerGroup.Group("/logs")
	{
		// Perhatikan ini menjadi POST "/", karena nanti dipanggil dari grup "/v1"
		logsRoutes.POST("/", h.ReceiveLog)
	}
}
