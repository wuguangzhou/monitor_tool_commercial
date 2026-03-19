package router

import (
	"backend/internal/handler"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func InitMonitorRouter(r *gin.Engine) {
	// 监控模块路由前缀，需与前端请求 `/api/monitor/...` 对齐
	monitorRouter := r.Group("/api/monitor")
	monitorRouter.Use(middleware.AuthMiddleware())
	{
		// 新增监控项
		monitorRouter.POST("/create", handler.CreateMonitorHandler)
		// 删除监控项
		monitorRouter.DELETE("/delete/:id", handler.DeleteMonitorHandler)
		// 更新监控项
		monitorRouter.PUT("/update/:id", handler.UpdateMonitorHandler)
		// 获取监控项列表（分页）
		monitorRouter.GET("/list", handler.GetMonitorListHandler)
		// 获取监控项详情+历史记录
		monitorRouter.GET("/detail/:id", handler.GetMonitorDetailHandler)
		// 手动执行一次监控检测
		monitorRouter.POST("/run/:id", handler.RunMonitorOnceHandler)
		// 暂停监控项
		monitorRouter.POST("/pause/:id", handler.PauseMonitorHandler)
		// 恢复监控项
		monitorRouter.POST("/resume/:id", handler.ResumeMonitorHandler)

	}
}
