package middleware

import (
	"hahacking/utils/acl"
	"hahacking/utils/cache"
	"hahacking/utils/response"
	"log"

	"github.com/gin-gonic/gin"
)

func ACL() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userName = c.GetHeader("x-requested-with")
		if userName == "" {
			response.Error(c, "header miss userName", nil)
			c.Abort()
			return
		}
		path := c.Request.URL.Path
		method := c.Request.Method
		cacheName := userName + "|" + path + "|" + method
		log.Println("cachename", cacheName)
		// 从缓存中读取&判断
		entry, err := cache.GlobalCache.Get(cacheName)
		log.Println("entry", entry)
		if err == nil && entry != nil {
			if string(entry) == "true" {
				c.Next()
			} else {
				response.Error(c, "access denied", nil)
				c.Abort()
				return
			}
		} else {
			// 从数据库中读取&判断
			//记录日志
			acl.Enforcer.EnableLog(true)
			// 加载策略规则
			err = acl.Enforcer.LoadPolicy()
			if err != nil {
				log.Println("loadPolicy error")
				panic(err)
			}
			// 验证策略规则

			result, err := acl.Enforcer.Enforce(userName, path, method)
			log.Println("result", result)
			if err != nil {
				response.Error(c, "No permission found", nil)
				c.Abort()
				return
			}
			if !result {
				// 添加到缓存中
				cache.GlobalCache.Set(cacheName, []byte("false"))
				response.Error(c, "access denied", nil)
				c.Abort()
				return
			} else {
				cache.GlobalCache.Set(cacheName, []byte("true"))
			}
			c.Next()
		}
	}
}
