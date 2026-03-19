package router

import "github.com/gin-gonic/gin"

func InitRouter() *gin.Engine {
	// 插件Gin引擎（开发环境使用gin.Debug()，生产环境使用gin.ReleaseModel）
	r := gin.Default()

	// 静态资源：头像等文件访问
	r.Static("/static", "./static")

	IintUserRouter(r)
	InitMonitorRouter(r)
	InitAlertRouter(r)

	return r
}
