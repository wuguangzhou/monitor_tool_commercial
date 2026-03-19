package router

import (
	"backend/internal/handler"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

// 初始化告警模块
func InitAlertRouter(r *gin.Engine) {
	alterRouter := r.Group("/api/alert")
	alterRouter.Use(middleware.AuthMiddleware())
	{
		alterRouter.POST("/config/update", handler.UpdateAlertConfigHandler)
		alterRouter.GET("/config", handler.GetAlertConfigHandler)
		alterRouter.GET("/list", handler.GetAlertListHandler)
	}
}
