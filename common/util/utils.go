package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GenCode() string {
	rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	str := fmt.Sprintf("%s%s%d", "MT", time.Now().Format("20060102"), rand1.Int())
	return str[:18]
}

func RandomString(length int) string {
	rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	str := fmt.Sprintf("%d", rand1.Int())
	return str[:length]
}

func GetQueryParams(c *gin.Context) map[string]interface{} {
	params := make(map[string]interface{})

	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}
	}
	return params
}

func ParseDateTime(s string, loc *time.Location, format string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if loc == nil {
		loc = time.Local
	}
	if format == "" {
		format = "2006-01-02 15:04"
	}
	t, err := time.ParseInLocation(format, s, loc)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
