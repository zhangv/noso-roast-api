package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// origin := c.Request.Header.Get("Origin")

		// if method == "OPTIONS" {
		// 	c.Header("Access-Control-Allow-Origin", "*")
		// 	c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
		// 	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		// 	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		// 	c.Header("Access-Control-Allow-Credentials", "true")
		// 	c.AbortWithStatus(http.StatusNoContent)
		// }

		// log.Println("22*******")
		// log.Println(method, origin)
		// log.Println("*******")
		// if origin != "" {
		// 	// c.Header("Access-Control-Allow-Origin", origin)
		// 	c.Header("Access-Control-Allow-Origin", "*")
		// 	c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
		// 	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		// 	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		// 	c.Header("Access-Control-Allow-Credentials", "true")

		// }

		// c.Next()

		// log.Println("333*******")
		// log.Println(method)
		// log.Println("*******")

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, token,X-Requested-With") //自定义 Header
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			// log.Println("aborting")
			c.AbortWithStatus(http.StatusNoContent)
		}
		// log.Println("???")
		c.Next()
	}

	// return cors.New(cors.Config{ //无效 - from origin 'http://localhost:9005' has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present on the requested resource.
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
	// 	AllowHeaders:     []string{"Origin, X-Requested-With,X-Requested-Device, X-Request-Platform, X-Request-Secret, X-Request-Version, Content-Type, Accept, Authorization, token, Access-Token"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	AllowOriginFunc: func(origin string) bool {
	// 		return true //origin == "http://localhost:9005"
	// 	},
	// 	MaxAge: 12 * time.Hour,
	// })
}
