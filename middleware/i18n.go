package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)
func I18n(bundle *i18n.Bundle) gin.HandlerFunc {
    return func(c *gin.Context) {
        lang := c.GetHeader("Accept-Language")
        if lang == "" {
            lang = "en"
        }
        localizer := i18n.NewLocalizer(bundle, lang)
        c.Set("localizer", localizer)
        c.Next()
    }
}