package routers

import (
	"hahacking/handler/websocket"
	"hahacking/i18n"
	"hahacking/middleware"
	"hahacking/routers/auth"
	"hahacking/routers/baidu"
	"hahacking/routers/bean"
	"hahacking/routers/category"
	"hahacking/routers/common"
	"hahacking/routers/cupping"
	"hahacking/routers/favorite"
	"hahacking/routers/order"
	"hahacking/routers/page"
	"hahacking/routers/price"
	"hahacking/routers/report"
	"hahacking/routers/roast"
	"hahacking/routers/shop"
	"hahacking/routers/supplier"
	"hahacking/routers/tag"
	"hahacking/routers/task"
	"hahacking/routers/user"
	"hahacking/version"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Page not found"})
	})
	r.Use(middleware.CustomLogger())
	// r.Use(gin.Logger()) //gin.Default()已经包含了gin.Logger()
	// r.Use(gin.Recovery()) //gin.Default中已经注册了gin.Recovery()
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(time.Second, 100, 100))
	r.Use(ErrorHandler())
	r.Use(i18n.Middleware())
	r.Use(middleware.QueryArrayCompatibility())
	RegisterAPI(r)
	RegisterWebSocket(r)
	return r
}
func RegisterWebSocket(app *gin.Engine) error {
	ws := app.Group("/ws")
	ws.GET("", func(c *gin.Context) {
		websocket.HandleConnections(c)
	})
	return nil
}
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(400, gin.H{
				"error": c.Errors.Last().Error(),
			})
		}
	}
}
func RegisterAPI(app *gin.Engine) error {
	apiV1 := app.Group("/")
	// apiV1.Use(middleware.JWTAuth(
	// 	"/auth/register",
	// 	"/auth/login",
	// 	"/auth/refresh", // 添加refresh路径到公开路径列表
	// ))
	baiduAPI := apiV1.Group("/baidu")
	{
		{
			baiduAPI.GET("text2audio/:text", baidu.Text2Audio)
		}
	}

	apiV1.GET("ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong@ api.noso.cafe", "time": time.Now().Format("2006-01-02 15:04:05"),
			"build_info": version.GetBuildInfo()})
	})

	common.RegisterRoutes(apiV1)
	cupping.RegisterRoutes(apiV1)
	auth.RegisterRoutes(apiV1)
	task.RegisterRoutes(apiV1)
	bean.RegisterRoutes(apiV1)
	roast.RegisterRoutes(apiV1)
	page.RegisterRoutes(apiV1)
	category.RegisterRoutes(apiV1)
	order.RegisterRoutes(apiV1)
	shop.RegisterRoutes(apiV1)
	user.RegisterRoutes(apiV1)
	tag.RegisterRoutes(apiV1)
	price.RegisterRoutes(apiV1)
	report.RegisterRoutes(apiV1)
	favorite.RegisterRoutes(apiV1)
	supplier.RegisterRoutes(apiV1)
	return nil

}
