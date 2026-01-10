package middleware

import (
	"hahacking/common/app"
	"hahacking/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 基础认证，支持给group配置后，限定到api一级的权限 - anyone, user, admin (参考userrole对象/表)
func BaseAuth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := app.Gin{Cxt: c}
		if len(roles) == 0 {
			c.Next()
			return
		}

		// 判断roles中是否包含"anyone"
		for _, role := range roles {
			if role == "anyone" {
				c.Next()
				return
			}
		}

		userrequired := false
		adminrequired := false
		for _, role := range roles {
			if role == "user" {
				userrequired = true
			}
			if role == "admin" {
				adminrequired = true
			}
		}
		if userrequired {
			uidStr := c.GetString("uid")
			uid, _ := strconv.Atoi(uidStr)
			if uid == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "请登录"})
				appG.Failed(http.StatusUnauthorized, "请登录")
				c.Abort()
				return
			} else {
				c.Next()
				return
			}
		}
		if adminrequired {
			uidStr := c.GetString("uid")
			uid, _ := strconv.Atoi(uidStr)
			userService := service.NewUserService()
			u, _ := userService.GetById(uid)
			uroles := u.GetRoles()
			// log.Println("roles", uroles)
			for _, role := range uroles {
				if role.Code == "admin" {
					c.Next()
					return
				}
			}
			// c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			appG.Failed(http.StatusForbidden, "无权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
