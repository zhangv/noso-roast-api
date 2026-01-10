package middleware

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// QueryArrayCompatibilityMiddleware 查询数组兼容性中间件
// 自动将 xxx[]=a&xxx[]=b 格式转换为 xxx=a&xxx=b 格式
func QueryArrayCompatibility() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取原始查询参数
		rawQuery := c.Request.URL.RawQuery
		if rawQuery == "" {
			c.Next()
			return
		}

		// 解析查询参数
		values, err := url.ParseQuery(rawQuery)
		if err != nil {
			c.Next()
			return
		}

		// 处理数组参数兼容性
		normalizedValues := make(url.Values)
		for key, valueList := range values {
			// 如果key以[]结尾，去掉[]并合并到普通key
			if strings.HasSuffix(key, "[]") {
				normalKey := strings.TrimSuffix(key, "[]")
				// 如果普通key已经存在，合并值
				if existing, exists := normalizedValues[normalKey]; exists {
					normalizedValues[normalKey] = append(existing, valueList...)
				} else {
					normalizedValues[normalKey] = valueList
				}
			} else {
				// 普通参数直接复制
				normalizedValues[key] = valueList
			}
		}

		// 重新构建查询字符串
		newQuery := normalizedValues.Encode()
		if newQuery != rawQuery {
			// 更新请求URL
			c.Request.URL.RawQuery = newQuery
		}

		c.Next()
	}
}
