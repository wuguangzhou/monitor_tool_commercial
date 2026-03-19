package router

import (
	"backend/internal/handler"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func IintUserRouter(r *gin.Engine) {
	// 用户路由组（无权限校验的接口）
	userRouter := r.Group("/api/user")
	{
		userRouter.POST("/register", handler.UserRegisterHandler)
		userRouter.POST("/login", handler.UserLoginHandler)

	}

	// 需要权限校验的用户路由组(添加auth中间件)
	authUserRouter := r.Group("/api/user")
	authUserRouter.Use(middleware.AuthMiddleware())
	{
		authUserRouter.GET("/info", handler.UserInfoHandler)
		authUserRouter.POST("/delete", handler.UserDeleteHandler)
		authUserRouter.POST("/avatar", handler.UploadAvatarHandler)
		authUserRouter.POST("/profile", handler.UpdateUserProfileHandler)
	}
}
