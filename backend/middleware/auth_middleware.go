package middleware

import (
	"backend/internal/handler"
	"backend/pkg/jwt"
	"backend/pkg/redis"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 权限校验中间件（验证用户是否登录）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头中的Authorization字段（格式：Bearer token）
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			handler.ResponseError(c, 401, "请先登录")
			c.Abort()
			return
		}
		//token := strings.TrimPrefix(authHeader, "Bearer ")
		// 校验HTTP请求头的Authotization字段是否为 Bearer+空格+Token的标准格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			handler.ResponseError(c, 401, "token格式错误")
			c.Abort()
			return
		}

		claim, err := jwt.ParseToken(parts[1])
		if err != nil {
			handler.ResponseError(c, 401, "token已过期或无效，请重新登录")
			c.Abort()
			return
		}

		// 校验redis中的token缓存
		cacheKey := "user:token:" + strconv.FormatInt(claim.UserId, 10)
		cacheToken, err := redis.Get(cacheKey)
		if err != nil || cacheToken != parts[1] {
			handler.ResponseError(c, 401, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set("userId", claim.UserId)
		c.Set("phone", claim.Phone)
		c.Next()
	}
}
