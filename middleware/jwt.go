package middleware

import (
	"hahacking/service"
	"hahacking/utils/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWT中间件，支持跳过公开路由
func JWTAuth(publicPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// // 检查当前路径是否在公开路径列表中 - 不需要了，改用BaseAuth。这里只检查jwt有效性及refresh
		currentPath := c.Request.URL.Path
		for _, path := range publicPaths {
			if strings.HasPrefix(currentPath, path) {
				c.Next() // 公开路由，直接放行
				return
			}
		}

		// 以下为受保护路由的JWT验证逻辑
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    response.ErrorCodeTokenInvalid,
				Message: "无效的令牌",
				Data:    nil,
			})
			c.Abort()
			return
		}
		userService := service.NewUserService()
		claims, err := userService.ParseJWT(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Response{
				Code:    response.ErrorCodeTokenExpired,
				Message: "令牌已过期",
				Data:    err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("uid", claims.Uid)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) (string, uint, bool) {
	username, ok1 := c.Get("username")
	uidStr, ok2 := c.Get("uid")
	if !ok1 || !ok2 {
		return "", 0, false
	}

	// Convert uid from string to uint
	uid, err := strconv.ParseUint(uidStr.(string), 10, 32)
	if err != nil {
		return "", 0, false
	}

	return username.(string), uint(uid), true
}
